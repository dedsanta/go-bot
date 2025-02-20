package web

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/service/lostfilm"
)

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

type Rss struct {
	XMLName struct{}   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RssChannel `xml:"channel"`
}

type RssChannel struct {
	Title         string           `xml:"title"`
	Link          string           `xml:"link"`
	LastBuildDate string           `xml:"lastBuildDate"`
	Items         []RssChannelItem `xml:"item"`
}

type RssChannelItem struct {
	Title        string `xml:"title"`
	Link         string `xml:"link"`
	PubDate      string `xml:"pubDate"`
	Description  string `xml:"description"`
	OriginalDate string `xml:"originalDate"`
	OriginalUrl  string `xml:"originalUrl"`
	Uid          string `xml:"uid"`
}

type LostFilmController struct {
	Service lostfilm.Service
}

func (c *LostFilmController) Add(g *gin.RouterGroup) {
	g.GET("/rss", c.rss())
}

// @Tags LostFilm controller
// @Param quality query string false "Quality filter"
// @Produce xml
// @Produce json
// @Success 200 {object} Rss
// @Failure 400,500 {object} HTTPError
// @Router /lostfilm/rss [get]
func (c *LostFilmController) rss() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		quality := ctx.Query("quality")
		episodes, err := c.Service.LastEpisodes(ctx)
		if err != nil {
			NewError(ctx, 500, err)
			return
		}
		rss := Rss{
			Version: "1.0",
			Channel: RssChannel{
				Title:         "Свежачок от LostFilm.TV",
				Link:          "https://www.lostfilm.tv/",
				LastBuildDate: episodes[0].Created.Format(dateLayout),
				Items:         make([]RssChannelItem, 0),
			},
		}

		for _, episode := range episodes {
			for _, file := range episode.ItemFiles {
				if quality != "" {
					if file.Quality != quality {
						continue
					}
				}
				description := fmt.Sprintf("%s\n<img src=\"%s\"/>", file.Description, episode.Poster)
				rss.Channel.Items = append(rss.Channel.Items, RssChannelItem{
					Title:        episode.Name + ". " + episode.EpisodeNameFull,
					Link:         config.GetConfig().Web.Domain + "/dl/" + file.GridFsId.Hex(),
					PubDate:      episode.Created.Format(dateLayout),
					Description:  description,
					OriginalDate: episode.Date.Format(dateLayout),
					OriginalUrl:  episode.Page,
					Uid:          episode.Id.Hex(),
				})
			}
		}

		ctx.XML(200, rss)
	}
}
