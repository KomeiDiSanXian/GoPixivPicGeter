package pixiv

import (
	"GoPixivPicGeter/crawler/pixiv/auth"
	"GoPixivPicGeter/model"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/tidwall/gjson"
)

var (
	Offset        = 30
	access_token  = ""
	refresh_token = ""
	proxyURL      = "http://127.0.0.1:10809/"
	Host          = "https://app-api.pixiv.net/v1" // need access token
	Illust        = "%s/illust/ranking?mode=%s&type=%s&offset=%d"
)

// mode
const (
	day          = "day"            // 日榜
	week         = "week"           // 周榜
	month        = "month"          // 月榜
	dayMale      = "day_male"       // 男性向日榜
	datFemale    = "day_female"     // 女性向日榜
	weekOriginal = "week_original"  // 原创周榜
	weekRookie   = "week_rookie"    // 新人周榜
	dayManga     = "day_manga"      // 漫画日榜
	dayR18       = "day_r18"        // r18日榜
	dayMaleR18   = "day_male_r18"   // 男性向r18日榜
	dayFemaleR18 = "day_female_r18" // 女性向r18日榜
	weekR18      = "week_r18"       //r18周榜
)

// illust type
const (
	illustration = "illust"
	comic        = "manga"
)

type Option struct {
	f func(*options)
}

type options struct {
	DisableKeepAlives,
	InsecureSkipVerify bool
	Proxy func(*http.Request) (*url.URL, error)
}

// WithDisableKeepAlives sets KeepAlives.
// Return Option which contains options pointer
func WithDisableKeepAlives(DisableKeepAlives bool) Option {
	return Option{func(o *options) {
		o.DisableKeepAlives = DisableKeepAlives
	},
	}
}

// WithInsecureSkipVerify.
// If true will skip certificate validation.
// Return Option which contains options pointer
func WithInsecureSkipVerify(InsecureSkipVerify bool) Option {
	return Option{func(o *options) {
		o.InsecureSkipVerify = InsecureSkipVerify
	}}
}

// WithProxy parses proxyURL and sets it as the proxy.
// Return Option which contains options pointer
func WithProxy(proxyURL string) Option {
	proxyAdress, _ := url.Parse(proxyURL)
	return Option{func(o *options) {
		o.Proxy = http.ProxyURL(proxyAdress)
	}}
}

// NewPixivClient will give a new *http.Client.
// Could be costomed by options.
// e.g. NewPixivClient(WithProxy(proxyURL))
func NewPixivClient(ops ...Option) *http.Client {
	opt := &options{}
	for _, op := range ops {
		op.f(opt)
	}

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: opt.InsecureSkipVerify,
	}
	tr.DisableKeepAlives = opt.DisableKeepAlives
	tr.Proxy = opt.Proxy
	return &http.Client{Transport: tr}
}

// NewPixivResp adds headers to disguise as pixiv app.
// Accept-Language is setted to "zh-cn".
// User-Agent is setted to "PixivIOSApp/7.13.3 (iOS 14.6; iPhone13,2)".
// app-version is 14.6.
// If client.Do(req) happens error, it will refresh token and retry the request 3 times
func NewPixivResp(method string, link string, body io.Reader) (*http.Response, error) {
	c := NewPixivClient(WithProxy(proxyURL))
	req, err := http.NewRequest(method, link, body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("app-os", "ios")
	req.Header.Add("app-os-version", "14.6")
	req.Header.Add("user-agent", "PixivIOSApp/7.13.3 (iOS 14.6; iPhone13,2)")
	req.Header.Add("Accept-Language", "zh-cn")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", access_token))
	var resp *http.Response
	for i := 0; i < 3; i++ {
		resp, err = c.Do(req)
		if resp.StatusCode == http.StatusBadRequest {
			access_token = auth.RefreshToken(refresh_token)
			continue
		}
		if err == nil {
			return resp, nil
		}
	}
	if err != nil {
		return nil, errors.New("[ERROR] Error NewPixivResp:" + err.Error())
	}
	return resp, nil
}

// GetPicRespBody splices a url with mode, illust_type and offset,
// then uses NewPixivResp() to get the response.
// Returns io.ReadAll(resp.Body)
func GetPicRespBody(mode, illust_type string, offset int) (respBody []byte, err error) {
	link := fmt.Sprintf(Illust, Host, mode, illust_type, offset)
	resp, err := NewPixivResp("GET", link, nil)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

// ReadIllustsFromBody reads body that comes from function GetPicRespBody().
// If picture is NSWF, isR18 should be true.
func ReadIllustsFromBody(body []byte, isR18 bool, illusts chan<- *model.Illust) {
	gjson.GetBytes(body, "illusts").ForEach(func(_, illust gjson.Result) bool {
		pic := illustJSONParse(illust, isR18)
		illusts <- pic
		return true
	})
}

// GetPicInfo returns *model.Illust array.
// mode comes from constants or one of this array:
// ["day","week","month","day_male","day_female","week_original","week_rookie","day_r18","day_male_r18","day_female_r18","week_r18"].
// n_offset is the number of the offset.
// If picture is came from day_r18，week_r18 and so on, isR18 should be true.
func GetPicInfo(mode string, n_offset int, isR18 bool) (illusts []*model.Illust) {
	illustsCh := make(chan *model.Illust)
	var wg sync.WaitGroup
	for i := 0; i < n_offset*Offset; i += Offset {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			log.Println("requesting offset", i)
			body, err := GetPicRespBody(mode, illustration, i)
			if err != nil {
				return
			}
			ReadIllustsFromBody(body, isR18, illustsCh)
		}(i)
	}

	go func() {
		wg.Wait()
		close(illustsCh)
	}()

	for illust := range illustsCh {
		illusts = append(illusts, illust)
	}

	return
}

// Parse illust JSON into model.Illust.
// Returns *model.Illust
func illustJSONParse(illust gjson.Result, isR18 bool) *model.Illust {
	var tags []model.Tag
	illust.Get("tags").ForEach(func(_, value gjson.Result) bool {
		tags = append(tags, model.Tag{Name: value.Get("name").Str, TransName: value.Get("translated_name").Str})
		return true
	})
	user := model.User{
		Account: illust.Get("user.account").Str,
		Name:    illust.Get("user.name").Str,
	}
	i := &model.Illust{
		Title:      illust.Get("title").Str,
		Type:       illust.Get("type").Str,
		Author:     user,
		AuthorID:   uint(illust.Get("user.id").Int()),
		IllustID:   uint(illust.Get("id").Uint()),
		UploadTime: illust.Get("create_date").Str,
		Tags:       tags,
		PageCount:  int(illust.Get("page_count").Int()),
		R18:        isR18,
	}
	return i
}
