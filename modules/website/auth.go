package website

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const steamLoginURL = "https://steamcommunity.com/openid/login"
const steamGetPlayerSummariesURL = "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/"

var loginUserURL = regexp.MustCompile(`^http://steamcommunity.com/openid/id/(\d+)$`)

type steamId64 int64

type steamOpenId struct {
	ns             string
	op_endpoint    string
	claimed_id     string
	identity       string
	return_to      string
	response_nonce string
	assoc_handle   string
	signed         string
	sig            string

	raw url.Values
}

type steamUser struct {
	SteamId                  string `json:"steamid"`
	CommunityVisibilityState int    `json:"communityvisibilitystate"`
	ProfileState             int    `json:"profilestate"`
	PersonaName              string `json:"personaname"`
	LastLogoff               int    `json:"lastlogoff"`
	ProfileURL               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	AvatarMedium             string `json:"avatarmedium"`
	AvatarFull               string `json:"avatarfull"`
	PersonaState             int    `json:"personastate"`
	PrimaryClanId            string `json:"primaryclanid"`
	TimeCreated              int    `json:"timecreated"`
	PersonaStateFlags        int    `json:"personastateflags"`
	LocCountryCode           string `json:"loccountrycode"`
	LocStateCode             string `json:"locstatecode"`
	LocCityId                int    `json:"loccityid"`
}

func parseOpenid(query url.Values) (*steamOpenId, bool) {
	var id steamOpenId
	id.ns = query.Get("openid.ns")
	id.op_endpoint = query.Get("openid.op_endpoint")
	id.claimed_id = query.Get("openid.claimed_id")
	id.identity = query.Get("openid.identity")
	id.return_to = query.Get("openid.return_to")
	id.response_nonce = query.Get("openid.response_nonce")
	id.assoc_handle = query.Get("openid.assoc_handle")
	id.signed = query.Get("openid.signed")
	id.sig = query.Get("openid.sig")
	id.raw = query

	if id.ns == "" {
		return nil, false
	}

	return &id, true
}

func validateOpenid(id *steamOpenId) (steamId64, error) {

	params := make(url.Values)
	params.Set("openid.assoc_handle", id.assoc_handle)
	params.Set("openid.signed", id.signed)
	params.Set("openid.sig", id.sig)
	params.Set("openid.ns", id.ns)

	signed := strings.Split(id.signed, ",")
	for _, item := range signed {
		params.Set("openid."+item, id.raw.Get("openid."+item))
	}
	params.Set("openid.mode", "check_authentication")

	validationResp, err := http.PostForm(steamLoginURL, params)
	if err != nil {
		return 0, err
	}
	defer validationResp.Body.Close()

	validationBody, err := ioutil.ReadAll(validationResp.Body)
	if err != nil {
		return 0, err
	}

	validationParts := strings.Split(string(validationBody), "\n")
	if err != nil {
		return 0, err
	}

	if len(validationParts) < 2 {
		return 0, errors.New("invalid validation result")
	}

	if validationParts[0] != "ns:"+id.ns {
		return 0, errors.New("ns mismatch")
	}

	if validationParts[1] != "is_valid:true" {
		return 0, errors.New("validation ailed")
	}

	matches := loginUserURL.FindStringSubmatch(id.claimed_id)
	if len(matches) != 2 {
		return 0, errors.New("invalid claimed_id: no steam id found")
	}

	steamId, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, errors.New("invalid claimed_id: invalid steam id size")
	}

	return steamId64(steamId), nil
}

func buildAuthUrl(callbackURL string) (string, error) {
	callback, err := url.Parse(callbackURL)
	if err != nil {
		return "", err
	}

	authUrl, err := url.Parse(steamLoginURL)
	if err != nil {
		return "", err
	}

	params := make(url.Values)
	params.Add("openid.ns", "http://specs.openid.net/auth/2.0")
	params.Add("openid.mode", "checkid_setup")
	params.Add("openid.return_to", callback.String())

	host := callback.Host
	portIndex := strings.Index(host, ":")
	if portIndex != -1 {
		host = string([]rune(host)[:portIndex])
	}

	params.Add("openid.realm", callback.Scheme+"://"+host)

	params.Add("openid.ns.sreg", "http://openid.net/extensions/sreg/1.1")
	params.Add("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
	params.Add("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	authUrl.RawQuery = params.Encode()

	return authUrl.String(), nil
}

func fetchUserProfile(steamKey string, steamId steamId64) (*steamUser, error) {
	params := make(url.Values)
	params.Set("key", steamKey)
	params.Set("steamids", strconv.FormatInt(int64(steamId), 10))
	playerSummaryURL := steamGetPlayerSummariesURL + "?" + params.Encode()

	profileResponse, err := http.Get(playerSummaryURL)
	if err != nil {
		return nil, err
	}

	defer profileResponse.Body.Close()

	body, err := ioutil.ReadAll(profileResponse.Body)
	if err != nil {
		return nil, err
	}

	type Result struct {
		Response struct {
			Players []steamUser `json:"players"`
		} `json:"response"`
	}

	var res Result
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if len(res.Response.Players) != 1 {
		return nil, errors.New("player profile not found")
	}

	p := res.Response.Players[0]
	return &p, nil
}
