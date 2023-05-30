package util

import (
	"image-service/core/domain"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

const (
	MinPageSize      = 20
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
	page, _ := strconv.Atoi(ParseQueryParam(req.URL.Query().Get("page")))
	perPage, _ := strconv.Atoi(ParseQueryParam(req.URL.Query().Get("perPage")))
	startDate, _ := strconv.Atoi(req.URL.Query().Get("startDate"))
	endDate, _ := strconv.Atoi(req.URL.Query().Get("endDate"))

	var filterData = domain.PageFilter{}

	filterData.Page = page
	filterData.PerPage = perPage
	filterData.StartDate = startDate
	filterData.EndDate = endDate

	if perPage == 0 {
		filterData.PerPage = MinPageSize
	}

	if page == 0 {
		filterData.Page = DefaultFirstPage
	}

	return filterData
}
