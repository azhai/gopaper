package common

import (
	"regexp"
	"strings"
)

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func ValidateDateFormat(date string) bool {
	if date == "" {
		return true
	}
	return dateRe.MatchString(date)
}

func ValidateTitle(title string) bool {
	return strings.TrimSpace(title) != "" && len(title) <= 200
}

func ValidateAuthor(author string) bool {
	return author == "" || len(author) <= 50
}

func ValidateTags(tags []string) bool {
	for _, tag := range tags {
		if len(tag) > 30 {
			return false
		}
	}
	return true
}

func ValidateArticleInput(dirPath, title, slug, author, date string, tags []string) map[string]string {
	errors := make(map[string]string)
	if strings.TrimSpace(dirPath) == "" {
		errors["dirPath"] = "目录路径不能为空"
	}
	if !ValidateTitle(title) {
		errors["title"] = "标题不能为空且不超过200字符"
	}
	if slug != "" && !ValidateSlug(slug) {
		errors["slug"] = "Slug仅允许小写字母、数字和连字符"
	}
	if !ValidateAuthor(author) {
		errors["author"] = "作者不超过50字符"
	}
	if !ValidateDateFormat(date) {
		errors["date"] = "日期格式错误，应为YYYY-MM-DD"
	}
	if !ValidateTags(tags) {
		errors["tags"] = "每个标签不超过30字符"
	}
	return errors
}
