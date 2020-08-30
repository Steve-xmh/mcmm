package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/steve-xmh/mcmm/twitchapi"
	"github.com/urfave/cli"
	"github.com/zhshch2002/goreq"
)

var (
	// Version 版本号
	Version = "0.0.1"
)

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Name = "Minecraft Mod Manager"
	app.Usage = "MCMM for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Authors = []cli.Author{
		{
			Name: "SteveXMH",
		},
	}
	app.Commands = []cli.Command{
		{
			// ==================== Search
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "search for a mod id by name",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "version, v",
					Value: "",
					Usage: "Choose the version of mod",
				},
			},
			Action: func(c *cli.Context) error {
				keyStr := c.Args().First()
				if keyStr == "" {
					cli.ShowCommandHelpAndExit(c, "search", 1)
					return nil
				}
				searchVer := c.String("version")
				if searchVer == "" {
					fmt.Println("Searching", keyStr, "for all Minecraft version")
				} else {
					fmt.Println("Searching", keyStr, "for Minecraft", searchVer)
				}
				mods, err := twitchapi.FindMods(keyStr, searchVer)
				if err != nil {
					return err
				}
				if len(mods) == 0 {
					fmt.Println("No mods found")
				}
				for _, mod := range mods {
					fmt.Println(mod.Name, "("+strconv.FormatInt(mod.Id, 10)+")", "-", mod.Summary)
				}
				return nil
			},
		}, {
			// ==================== Download
			Name:    "download",
			Aliases: []string{"d"},
			Usage:   "Download a mod with dependcises from curseforge",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "version, v",
					Value: "",
					Usage: "Choose the version of mod",
				},
			},
			Action: func(c *cli.Context) error {
				wg := sync.WaitGroup{}
				down := func(idStr string) error {
					defer wg.Done()
					if idStr == "" {
						cli.ShowCommandHelpAndExit(c, "download", 1)
						return nil
					}
					installVer := c.String("version")
					if installVer == "" {
						fmt.Println("Getting", idStr, "for latest Minecraft version")
					} else {
						fmt.Println("Getting", idStr, "for Minecraft", installVer)
					}
					var mod twitchapi.ModFile
					var err error
					idInt, err := strconv.ParseInt(idStr, 10, 64)
					if err != nil {
						mods, err := twitchapi.FindMods(idStr, installVer)
						if err != nil {
							fmt.Println("Error on getting mod info:", err)
							return err
						}
						if len(mods) == 0 {
							fmt.Println("Can't find any mod called", idStr)
							return nil
						}
						mod, err = twitchapi.GetModFileUrl(mods[0].Id, installVer)
						if err != nil {
							fmt.Println("Error on getting file url:", err)
							return err
						}
					} else {
						mod, err = twitchapi.GetModFileUrl(idInt, installVer)
						if err != nil {
							fmt.Println("Error on getting file url:", err)
							return err
						}
					}
					o, err := os.Create(mod.FileName)
					if err != nil {
						fmt.Println("Error on open save file:", err)
						return err
					}
					fmt.Println("Downloading", mod.FileName, ":", mod.Url)
					var size int = 0
					if res, err := goreq.Do(goreq.Get(mod.Url)).Resp(); err != nil {
						fmt.Println("Error on downloading file:", err)
						return err
					} else {
						writeLen, err := o.Write(res.Body)
						size += writeLen
						if err != nil {
							fmt.Println("Error on writing save file:", err)
							return err
						}
					}
					fmt.Println("Saved mod", mod.FileName, "("+strconv.Itoa(size)+" bytes)")
					err = o.Close()
					if err != nil {
						fmt.Println("Error on defer close file:", err)
						return err
					}
					return nil
				}
				for i := 0; i < c.NArg(); i++ {
					wg.Add(1)
					go down(c.Args().Get(i))
				}
				wg.Wait()
				return nil
			},
		}, {
			// ==================== Info
			Name:  "info",
			Usage: "Get mod info and version compatibility list.",
			Action: func(c *cli.Context) error {
				idStr := c.Args().First()
				if c.Args().First() == "" {
					cli.ShowCommandHelpAndExit(c, "info", 1)
					return nil
				}
				var mod twitchapi.ModInfo
				var err error
				idInt, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					mods, err := twitchapi.FindMods(idStr, "")
					if err != nil {
						fmt.Println("Error on getting mod info:", err)
						return err
					}
					if len(mods) == 0 {
						fmt.Println("Can't find any mod called", idStr)
						return nil
					}
					mod = mods[0]
				} else {
					mod, err = twitchapi.GetMod(idInt)
					if err != nil {
						fmt.Println("Error on getting file url:", err)
						return err
					}
				}
				fmt.Println(mod.Id, "-", mod.Name)
				fmt.Print("Supported versions: ")
				for _, ver := range mod.SupportVersions {
					fmt.Print(ver)
					fmt.Print(" ")
				}
				if len(mod.SupportVersions) == 0 {
					fmt.Print("Nothing\n")
				} else {
					fmt.Print("\n")
				}
				fmt.Println(mod.Summary)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
