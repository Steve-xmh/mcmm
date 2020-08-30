package twitchapi

import (
	"errors"
	"strconv"

	"github.com/zhshch2002/goreq"
)

var (
	apiBaseURL = "https://addons-ecs.forgesvc.net"
)

type ModFile struct {
	Id       int64
	FileName string
	Url      string
}

type ModInfo struct {
	Id              int64
	Name            string
	Files           []ModFile
	Summary         string
	SupportVersions []string
}

func GetMod(modId int64) (ModInfo, error) {
	var err error
	var searchUrl = apiBaseURL + "/api/v2/addon/" + strconv.FormatInt(modId, 10)
	req := goreq.Get(searchUrl)
	modJson, err := req.Do().JSON()
	if err != nil {
		return ModInfo{}, err
	}
	var result ModInfo
	var files []ModFile
	for _, fileJson := range modJson.Get("latestFiles").Array() {
		files = append(files, ModFile{
			Id:       fileJson.Get("id").Int(),
			FileName: fileJson.Get("fileName").String(),
			Url:      fileJson.Get("downloadUrl").String(),
		})
	}
	var supportVersions []string
	for _, versionJson := range modJson.Get("gameVersionLatestFiles").Array() {
		supportVersions = append(supportVersions, versionJson.Get("gameVersion").String())
	}
	result = ModInfo{
		Id:              modJson.Get("id").Int(),
		Name:            modJson.Get("name").String(),
		Files:           files,
		SupportVersions: supportVersions,
		Summary:         modJson.Get("summary").String(),
	}
	return result, nil
}

func FindMods(searchFilter string, mcVersion string) ([]ModInfo, error) {
	var err error
	var searchUrl = apiBaseURL + "/api/v2/addon/search?gameId=432&sectionId=6&"
	if mcVersion != "" {
		searchUrl += "gameVersion=" + mcVersion + "&"
	}
	searchUrl += "searchFilter=" + searchFilter
	req := goreq.Get(searchUrl)
	json, err := req.Do().JSON()
	if err != nil {
		return nil, err
	}
	modArrJson := json.Array()
	var results []ModInfo
	for _, modJson := range modArrJson {
		var files []ModFile
		for _, fileJson := range modJson.Get("latestFiles").Array() {
			files = append(files, ModFile{
				Id:       fileJson.Get("id").Int(),
				FileName: fileJson.Get("fileName").String(),
				Url:      fileJson.Get("downloadUrl").String(),
			})
		}
		var supportVersions []string
		for _, versionJson := range modJson.Get("gameVersionLatestFiles").Array() {
			supportVersions = append(supportVersions, versionJson.Get("gameVersion").String())
		}
		results = append(results, ModInfo{
			Id:              modJson.Get("id").Int(),
			Name:            modJson.Get("name").String(),
			Files:           files,
			SupportVersions: supportVersions,
			Summary:         modJson.Get("summary").String(),
		})
	}
	return results, nil
}

func GetModFileUrl(id int64, mcVersion string) (ModFile, error) {
	var searchUrl = apiBaseURL + "/api/v2/addon/" + strconv.FormatInt(id, 10)
	filesReq := goreq.Get(searchUrl + "/files")
	json, err := filesReq.Do().JSON()
	if err != nil {
		return ModFile{}, err
	}
	req := goreq.Get(searchUrl)
	infoJson, err := req.Do().JSON()
	if err != nil {
		return ModFile{}, err
	}
	if mcVersion == "" {
		// Search default file id
		defaultFileId := infoJson.Get("defaultFileId").Int()
		for _, file := range json.Array() {
			if file.Get("id").Int() == defaultFileId {
				return ModFile{
					FileName: file.Get("fileName").String(),
					Id:       file.Get("id").Int(),
					Url:      file.Get("downloadUrl").String(),
				}, nil
			}
		}
	} else {
		for _, file := range json.Array() {
			for _, version := range file.Get("gameVersion").Array() {
				if version.String() == mcVersion {
					return ModFile{
						FileName: file.Get("fileName").String(),
						Id:       file.Get("id").Int(),
						Url:      file.Get("downloadUrl").String(),
					}, nil
				}
			}
		}
	}
	return ModFile{}, errors.New("Can't found " + infoJson.Get("name").String() + " (" + strconv.FormatInt(id, 10) + ") in version " + mcVersion)
}
