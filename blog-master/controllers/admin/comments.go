package admin

import (
	"blog-master/models"
	"strings"
	"strconv"
	"blog-master/controllers/ipfilter"
)

type CommentsController struct {
	baseController
}

//评论列表
func (this *CommentsController) List() {
	if this.userid != 1 {
		this.showmsg("未授权访问")
	}
	var list []*models.Comments
	var comment models.Comments
	var (
		page       int64
		pagesize   int64 = 10
		offset     int64
	)
	if page, _ = this.GetInt64("page"); page < 1 {
		page = 1
	}
	offset = (page - 1) * pagesize
	count, _ := comment.Query().Count()
	if count > 0 {
		comment.Query().OrderBy("-submittime").Limit(pagesize, offset).All(&list)
	}
	this.Data["pagebar"] = models.NewPager(page, count, pagesize, "/admin/comments/list?page=%d").ToString()
	this.Data["list"] = list
	this.display()
}

//添加评论
func (this *CommentsController) Add() {
	x := ipfilter.ConnFilterCtx().GetabnConn(this.clientip)
	if x > 0 {
		this.Abort("500")
	} else {
		if this.Ctx.Request.Method == "POST" {
			var comment models.Comments
			blogid, _ := strconv.Atoi(strings.TrimSpace(this.GetString("object_pk")))
			replypk := strings.TrimSpace(this.GetString("reply_pk"))
			replyfk, _ := strconv.Atoi(strings.TrimSpace(this.GetString("reply_fk")))
			comment_content := strings.TrimSpace(this.GetString("comment_content"))
			security_hash := strings.TrimSpace(this.GetString("security_hash"))
			timestamp := strings.TrimSpace(this.GetString("timestamp"))
			if comment_content != "" && security_hash != "" {
				checkstr := models.Md5([]byte(replypk + timestamp + "@YO!r52w!D2*I%Ov"))
				//println(security_hash,checkstr)
				if checkstr == security_hash {
					comment.Comment = comment_content
					var user models.User
					user.Query().Filter("id", this.userid).Limit(1).One(&user)
					comment.User = &models.User{Id: this.userid}
					comment.Obj_pk = &models.Post{Id: int64(blogid)}
					replypk_to_int, _ := strconv.Atoi(replypk)
					comment.Reply_pk = int64(replypk_to_int)
					comment.Reply_fk = int64(replyfk)
					comment.Ipaddress = this.getClientIp()
					comment.Submittime = this.getTime()
					models.Cache.Delete("newcomments")
					if err := comment.Insert(); err != nil {
						this.showmsg(err.Error())
					}
				}
			}
			this.Redirect(this.Ctx.Request.Referer(), 302)
		} else {
			this.Abort("404")
		}
	}
}

//编辑评论
func (this *CommentsController) Edit() {
	if this.userid != 1 {
		this.showmsg("未授权操作")
	}
	id, _ := this.GetInt64("id")
	comment := models.Comments{Id: id}
	if comment.Read() != nil {
		this.showmsg("未查询到该评论")
	}
	if this.Ctx.Request.Method == "POST" {
		content := strings.TrimSpace(this.GetString("content"))
		is_removed, _ := this.GetInt8("is_removed")
		comment.Comment = content
		comment.Is_removed = is_removed
		comment.Update("comment", "is_removed")
		models.Cache.Delete("newcomments")
		this.Redirect("/admin/comments/list", 302)
	}
	this.Data["comment"] = comment
	this.display()
}

//删除评论
func (this *CommentsController) Delete() {
	if this.userid != 1 {
		this.showmsg("未授权操作")
	}
	id, _ := this.GetInt64("id")
	comment := models.Comments{Id: id}
	if comment.Read() != nil {
		this.showmsg("未查询到该评论")
	}
	comment.Is_removed = 1
	comment.Update("is_removed")
	models.Cache.Delete("newcomments")
	this.Redirect("/admin/comments/list", 302)
}
