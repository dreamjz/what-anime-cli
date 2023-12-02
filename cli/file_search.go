package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	helpers "github.com/irevenko/what-anime-cli/helpers"
	types "github.com/irevenko/what-anime-cli/types"
	"github.com/muesli/termenv"
)

const (
	fileSearchURL = "https://api.trace.moe/search?anilistInfo"
)

// SearchByImageFile is for finding the anime scene by existing image file
func SearchByImageFile(imagePath string) {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) { // 检查文件是否存在
		if err != nil {
			log.Fatal("Invalid file path")
		}
	}

	termenv.HideCursor() // 隐藏 cursor
	defer termenv.ShowCursor()

	s := spinner.New(spinner.CharSets[33], 100*time.Millisecond) // 新建进度条
	s.Prefix = "🔎 Searching for the anime: "                     // 进度条前置信息
	s.FinalMSG = color.GreenString("✔️  Found!\n")               // 进度条完成信息

	go catchInterrupt(s) // 收到 SIGINT 时优雅的结束

	s.Start() // 启动进度条

	imageFile, err := os.Open(imagePath) // 打开图片文件
	helpers.HandleError(err)             // 处理错误

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part, _ := writer.CreateFormFile("image", filepath.Base(imagePath)) // 添加 form-data

	_, err = io.Copy(part, imageFile) // 将文件写入 form-data
	helpers.HandleError(err)

	err = writer.Close()
	helpers.HandleError(err)

	resp, err := http.Post(fileSearchURL, writer.FormDataContentType(), payload) // 发送 POST 请求
	helpers.HandleError(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body) // 读取返回信息
	helpers.HandleError(err)

	var animeResp types.Response
	json.Unmarshal(body, &animeResp)

	s.Stop() // 进度条结束

	fmt.Println("🌸 Title Native:", animeResp.Result[0].Anilist.Title.Native)
	fmt.Println("🗻 Title Romaji:", animeResp.Result[0].Anilist.Title.Romaji)
	fmt.Println("🗽 Title English:", animeResp.Result[0].Anilist.Title.English)
	fmt.Print("📊 Similarity: ")
	helpers.PrintAnimeSimilarity(strconv.FormatFloat(animeResp.Result[0].Similarity, 'f', 6, 64))
	_, err = fmt.Fprintln(color.Output, "📺 Episode Number: "+color.MagentaString(strconv.Itoa(animeResp.Result[0].Episode)))
	helpers.HandleError(err)
	fmt.Print("⌚ Scene From: ")
	helpers.PrintSceneTime(animeResp.Result[0].From)
	fmt.Print("⌚ Scene To: ")
	helpers.PrintSceneTime(animeResp.Result[0].To)
	fmt.Print("🍓 Is Adult: ")
	helpers.PrintIsAdult(animeResp.Result[0].Anilist.IsAdult)
	//fmt.Println(string(body))
}
