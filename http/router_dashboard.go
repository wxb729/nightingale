package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/didi/nightingale/v5/models"
)

func dashboardGets(c *gin.Context) {
	limit := queryInt(c, "limit", defaultLimit)
	query := queryStr(c, "query", "")
	onlyfavorite := queryBool(c, "onlyfavorite", false)

	me := loginUser(c)
	ids, err := me.FavoriteDashboardIds()
	dangerous(err)

	// 我的收藏是空的，所以直接返回空列表
	if onlyfavorite && len(ids) == 0 {
		renderZeroPage(c)
		return
	}

	total, err := models.DashboardTotal(onlyfavorite, ids, query)
	dangerous(err)

	list, err := models.DashboardGets(onlyfavorite, ids, query, limit, offset(c, limit))
	dangerous(err)

	if onlyfavorite {
		for i := 0; i < len(list); i++ {
			list[i].Favorite = 1
		}
	} else {
		for i := 0; i < len(list); i++ {
			list[i].FillFavorite(ids)
		}
	}

	renderData(c, gin.H{
		"list":  list,
		"total": total,
	}, nil)
}

func dashboardGet(c *gin.Context) {
	renderData(c, Dashboard(urlParamInt64(c, "id")), nil)
}

type dashboardForm struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Tags    string `json:"tags"`
	Configs string `json:"configs"`
}

func dashboardAdd(c *gin.Context) {
	var f dashboardForm
	bind(c, &f)

	me := loginUser(c).MustPerm("dashboard_create")

	d := &models.Dashboard{
		Name:     f.Name,
		Tags:     f.Tags,
		Configs:  f.Configs,
		CreateBy: me.Username,
		UpdateBy: me.Username,
	}

	dangerous(d.Add())

	renderData(c, d, nil)
}

func dashboardPut(c *gin.Context) {
	var f dashboardForm
	bind(c, &f)

	me := loginUser(c).MustPerm("dashboard_modify")
	d := Dashboard(urlParamInt64(c, "id"))

	if d.Name != f.Name {
		num, err := models.DashboardCount("name=? and id<>?", f.Name, d.Id)
		dangerous(err)

		if num > 0 {
			bomb(200, "Dashboard %s already exists", f.Name)
		}
	}

	d.Name = f.Name
	d.Tags = f.Tags
	d.Configs = f.Configs
	d.UpdateAt = time.Now().Unix()
	d.UpdateBy = me.Username

	dangerous(d.Update("name", "tags", "configs", "update_at", "update_by"))

	renderData(c, d, nil)
}

func dashboardClone(c *gin.Context) {
	var f dashboardForm
	bind(c, &f)

	me := loginUser(c).MustPerm("dashboard_create")

	d := &models.Dashboard{
		Name:     f.Name,
		Tags:     f.Tags,
		Configs:  f.Configs,
		CreateBy: me.Username,
		UpdateBy: me.Username,
	}
	dangerous(d.AddOnly())

	chartGroups, err := models.ChartGroupGets(f.Id)
	dangerous(err)
	for _, chartGroup := range chartGroups {
		charts, err := models.ChartGets(chartGroup.Id)
		dangerous(err)
		chartGroup.DashboardId = d.Id
		chartGroup.Id = 0
		dangerous(chartGroup.Add())

		for _, chart := range charts {
			chart.Id = 0
			chart.GroupId = chartGroup.Id
			dangerous(chart.Add())
		}
	}

	renderData(c, d, nil)
}

func dashboardDel(c *gin.Context) {
	loginUser(c).MustPerm("dashboard_delete")
	renderMessage(c, Dashboard(urlParamInt64(c, "id")).Del())
}

func dashboardFavoriteAdd(c *gin.Context) {
	me := loginUser(c)
	d := Dashboard(urlParamInt64(c, "id"))
	renderMessage(c, models.DashboardFavoriteAdd(d.Id, me.Id))
}

func dashboardFavoriteDel(c *gin.Context) {
	me := loginUser(c)
	d := Dashboard(urlParamInt64(c, "id"))
	renderMessage(c, models.DashboardFavoriteDel(d.Id, me.Id))
}
