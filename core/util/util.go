package util

import (
	"image-service/core/domain"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

const (
	MinPageSize      = 15
	DefaultFirstPage = 1
)

func ParseQueryParam(param string) string {
	reg, err := regexp.Compile(`[!?;{}<>%'=]`)
	if err != nil {
		log.Println(err)
	}
	res := reg.ReplaceAllString(param, "")
	return res
}

func PageFilter(req *http.Request) domain.PageFilter {
	page, err := strconv.Atoi(ParseQueryParam(req.URL.Query().Get("page")))
	if err != nil {
		log.Println(err)
	}
	perPage, err := strconv.Atoi(ParseQueryParam(req.URL.Query().Get("per_page")))
	if err != nil {
		log.Println(err)
	}

	var filterData = domain.PageFilter{}

	if perPage == 0 {
		filterData.PerPage = MinPageSize
	}

	if page == 0 {
		filterData.Page = DefaultFirstPage
	}
	filterData.Page = page
	filterData.PerPage = perPage

	return filterData
}
