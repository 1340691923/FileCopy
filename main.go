package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type FileInfo struct {
	Houzhui string
	FileName string
	FileNames string
	Dir      string
}

type MyWindow struct {
	*walk.MainWindow
	TxtText      *walk.LineEdit
	SourceDirText     *walk.LineEdit
	TargetDirText     *walk.LineEdit
	AllCountText 	*walk.LineEdit
	TxtCountText    *walk.LineEdit
	DealCountText    *walk.LineEdit
	NeedCountText   *walk.LineEdit
	ExecBtn *walk.PushButton
	CopeBtn *walk.RadioButton
	CutBtn  *walk.RadioButton
	ChooseTxtBtn *walk.PushButton
	ChooseSourceBtn *walk.PushButton
	ChooseTargetBtn *walk.PushButton
	ChooseExportBtn *walk.PushButton
	Baz int
	//存放源文件夹 找到的文件
	FindArr []string
	ExportBtn *walk.PushButton
	lb *walk.ListBox
	ExportTxt  *walk.LineEdit
	DumpBtn *walk.RadioButton
	NoDumpBtn  *walk.RadioButton
	Dump int
	TxtDealCountText *walk.LineEdit
	StopBtn *walk.PushButton
	fb *walk.ListBox
	lineMap map[string]string
}

var lineSlice []string
var wg sync.WaitGroup
var fileWg sync.WaitGroup
var err error
var lock sync.Mutex
var ch = make(chan struct{}, 255)
var stopFlag bool = false
var needMoveCount int
type Charset string

const(
	TXTFILEPATH int32 = 1
	EXPORTFILEPATH int32 = 2
	SOURCEDIRPATH int32 = 1
	TARGETDIRPATH int32 = 2
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
	COPY = 1
	CUT = 2
	DUMP = 1
	NODUMP = 2
)

func main(){
	mw := new(MyWindow)
	mw.Baz = COPY
	mw.Dump = DUMP
	if err := (MainWindow{
		Icon:"icon.ico",
		Font:Font{Family:"Microsoft YaHei",PointSize:8,Bold:true},
		AssignTo: &mw.MainWindow,
		Title:    "挑歌拷贝go语言程序 by 肖文龙  微信 :qq1340691923",
		MinSize:  Size{350, 450},
		Size:Size{700,700},
		Layout:   VBox{},
		DataBinder: DataBinder{
			DataSource: mw,
			AutoSubmit: true,
			OnSubmitted: func() {},
		},
		OnDropFiles: func(files []string) {
			if mw.TxtText.Text()==""{
				mw.TxtText.SetText(strings.Join(files, "\r\n"))
			}else if mw.SourceDirText.Text()==""{
				mw.SourceDirText.SetText(strings.Join(files, "\r\n"))
			}else if mw.TargetDirText.Text()==""{
				mw.TargetDirText.SetText(strings.Join(files, "\r\n"))
			}else if mw.ExportTxt.Text()==""{
				mw.ExportTxt.SetText(strings.Join(files, "\r\n"))
			}
		},
		Children: []Widget{
			Composite{
				Layout:  HBox{Margins:Margins{Top:10}},
				Children: []Widget{
					GroupBox{
						Title:  "执行条件",
						Layout: Grid{Columns:3,},
						Children: []Widget{
							Label{Text: "配置文件:"},
							LineEdit{AssignTo: &mw.TxtText},
							PushButton{
								AssignTo: &mw.ChooseTxtBtn,
								Text:     "选择配置文件",
							},
							Label{Text: "源文件夹:"},
							LineEdit{AssignTo: &mw.SourceDirText},
							PushButton{
								AssignTo: &mw.ChooseSourceBtn,
								Text:     "选择源文件夹",
							},
							Label{Text: "目标文件夹:"},
							LineEdit{AssignTo: &mw.TargetDirText},
							PushButton{
								AssignTo: &mw.ChooseTargetBtn,
								Text:     "选择目标文件夹",
							},
							Label{Text: "导出文件位置:"},
							LineEdit{AssignTo: &mw.ExportTxt},
							PushButton{
								AssignTo: &mw.ChooseExportBtn,
								Text:     "选择导出文件位置",
							},
							RadioButtonGroup{
								DataMember: "Dump",
								Buttons: []RadioButton{
									RadioButton{
										Name:  "Dump",
										Text:  "跳过",
										Value: DUMP,
										AssignTo:&mw.DumpBtn,
									},
									RadioButton{
										Name:  "NoDump",
										Text:  "覆盖",
										Value: NODUMP,
										AssignTo:&mw.NoDumpBtn,
									},
								},
							},
						},
					},
				},
			},
			Composite{
				Layout: Grid{Columns: 5, Spacing: 20},
				Children: []Widget{
					RadioButtonGroup{
						DataMember: "Baz",
						Buttons: []RadioButton{
							RadioButton{
								Name:  "copy",
								Text:  "复制",
								Value: COPY,
								AssignTo:&mw.CopeBtn,
							},
							RadioButton{
								Name:  "cut",
								Text:  "剪切",
								Value: CUT,
								AssignTo:&mw.CutBtn,
							},
						},
					},
					PushButton{
						AssignTo: &mw.ExecBtn,
						Text:     "执行",
					},
					PushButton{
						AssignTo: &mw.StopBtn,
						Text:     "暂停",
						Enabled:false,
						OnClicked: func() {
							go mw.StopAction()
						},
					},
					PushButton{
						AssignTo: &mw.ExportBtn,
						Text:     "导出",
						Enabled:false,
					},
				},
			},
			Composite{
				Layout: Grid{Columns: 2, Spacing: 80},
				Children: []Widget{
					ListBox{
						AssignTo: &mw.lb,
					},
					ListBox{
						AssignTo: &mw.fb,
						MinSize:Size{Width:300},
					},
				},
			},
			Composite{
				Layout:  HBox{},
				Children: []Widget{
					GroupBox{
						Title:  "执行结果",
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Label{Text: "配置文件已完成数量:"},
							LineEdit{AssignTo: &mw.TxtDealCountText, ReadOnly: true},
							Label{Text: "配置文件内的配置数量:"},
							LineEdit{AssignTo: &mw.TxtCountText, ReadOnly: true},
							Label{Text: "从源文件夹检索到的文件总数量:"},
							LineEdit{AssignTo: &mw.AllCountText, ReadOnly: true},
							Label{Text: "从源文件夹检索到的匹配的拷贝文件数量:"},
							LineEdit{AssignTo: &mw.NeedCountText, ReadOnly: true},
							Label{Text: "当前拷贝完的文件数量:"},
							LineEdit{AssignTo: &mw.DealCountText, ReadOnly: true},
						},
					},
				},
			},
		},
	}).Create(); err != nil {
		log.Fatalln(err)
	}

	mw.ExecBtn.Clicked().Attach(func() {go mw.ExecAction()})
	mw.ExportBtn.Clicked().Attach(func() {go mw.ExportAction()})
	mw.ChooseTxtBtn.Clicked().Attach(func() {go mw.OpenFileActionTriggered(TXTFILEPATH)})
	mw.ChooseExportBtn.Clicked().Attach(func() {go mw.OpenFileActionTriggered(EXPORTFILEPATH)})
	mw.ChooseSourceBtn.Clicked().Attach(func() {go mw.OpenDirActionTriggered(SOURCEDIRPATH)})
	mw.ChooseTargetBtn.Clicked().Attach(func() {go mw.OpenDirActionTriggered(TARGETDIRPATH)})
	mw.Run()
}

//递归计算目录下所有文件
func(mw *MyWindow)WalkDir(path string, filePath chan <- string){
	ch <- struct{}{} //限制并发量
	entries, _ := ioutil.ReadDir(path)
	<- ch
	for _, e := range entries{
		if e.IsDir() {
			fileWg.Add(1)
			go mw.WalkDir(filepath.Join(path, e.Name()), filePath)
		} else {
			filePath <- path+"\\"+e.Name()
		}
	}
	defer fileWg.Done()
}

//暂停
func(mw *MyWindow)StopAction(){
	mw.Error("暂停成功")
	mw.StopBtn.SetEnabled(false)
	stopFlag = true
}

//导出未完成的配置
func(mw *MyWindow)ExportAction(){
	defer mw.Close()
	mw.ExportBtn.SetEnabled(false)
	txtText := mw.TxtText.Text()
	sourceDirText := mw.SourceDirText.Text()
	targetDirText := mw.TargetDirText.Text()
	exportTxt := mw.ExportTxt.Text()
	if txtText == ""{
		mw.Error("配置文件地址不能为空")
		return
	}
	if sourceDirText == ""{
		mw.Error("源文件夹地址不能为空")
		return
	}
	if targetDirText == ""{
		mw.Error("目标文件夹地址不能为空")
		return
	}
	if exportTxt == ""{
		mw.Error("导出文件地址不能为空")
		return
	}

	f, err := os.Create(exportTxt)

	if err != nil {
		mw.Error("创建文件失败："+err.Error())
		return
	}

	w := bufio.NewWriter(f)
	for f,_ := range mw.lineMap{
		wg.Add(1)
		go func(f string){
			lock.Lock()
			lineStr := fmt.Sprintf("%s", f)
			fmt.Fprintln(w, lineStr)
			lock.Unlock()
			wg.Done()
		}(f)
	}
	wg.Wait()
	w.Flush()
	f.Close()
	mw.Success("文件导出完毕，文件地址为："+exportTxt)
}

//选择文件夹
func (mw *MyWindow)OpenDirActionTriggered(types int32){
	dlg := new(walk.FileDialog)
	dlg.Title = "打开文件"
	dlg.Filter = "文本文件 (*.txt)|*.txt|所有文件 (*.*)|*.*"
	if ok, err := dlg.ShowBrowseFolder(mw); err != nil {
		mw.Error("选择文件时异常!")
		return
	} else if !ok {
		return
	}
	s := fmt.Sprintf("%s", dlg.FilePath)
	if types == SOURCEDIRPATH{
		mw.SourceDirText.SetText(s)
	}else if types == TARGETDIRPATH{
		mw.TargetDirText.SetText(s)
	}
}

//选择文件
func (mw *MyWindow)OpenFileActionTriggered(types int32){
	dlg := new(walk.FileDialog)
	dlg.Title = "打开文件"
	dlg.Filter = "文本文件 (*.txt)|*.txt|所有文件 (*.*)|*.*"
	if ok, err := dlg.ShowOpen(mw); err != nil {
		mw.Error("选择文件时异常!")
		return
	} else if !ok {
		return
	}

	s := fmt.Sprintf("%s", dlg.FilePath)
	if types == TXTFILEPATH{
		mw.TxtText.SetText(s)
	}else if types == EXPORTFILEPATH{
		mw.ExportTxt.SetText(s)
	}
}

//执行
func (mw *MyWindow) ExecAction() {
	defer mw.Close()
	runtime.GOMAXPROCS(runtime.NumCPU())


	mw.ExecBtn.SetEnabled(false)
	mw.ExecBtn.SetText("执行中...")
	txtText := mw.TxtText.Text()
	sourceDirText := mw.SourceDirText.Text()
	targetDirText := mw.TargetDirText.Text()

	if txtText == ""{
		mw.Error("配置文件地址不能为空")
		return
	}

	if sourceDirText == ""{
		mw.Error("源文件夹地址不能为空")
		return
	}

	if targetDirText == ""{
		mw.Error("目标文件夹地址不能为空")
		return
	}

	flag,err := mw.PathExists(txtText)
	if err!=nil{
		mw.Error(err.Error())
		return
	}
	flag2,err := mw.PathExists(sourceDirText)
	if err!=nil{
		mw.Error(err.Error())
		return
	}
	flag3,err := mw.PathExists(targetDirText)
	if err!=nil{
		mw.Error(err.Error())
		return
	}

	if err!=nil{
		mw.Error(err.Error())
		return
	}
	if !flag{
		mw.Error("配置文件地址不存在，请检查")
		return
	}
	if !flag2{
		mw.Error("读取文件夹地址不存在，请检查")
		return
	}
	if !flag3{
		mw.Error("目的文件夹地址不存在，请检查")
		return
	}
	if !stopFlag{
		mw.TxtDealCountText.SetText("0")
		mw.DealCountText.SetText("0")
		mw.lineMap = make(map[string]string,0)
		err = mw.ReadLineFile(txtText)
	}
	//配置文件列表 赋值
	mw.lb.SetModel(lineSlice)
	mw.TxtCountText.SetText(strconv.Itoa(len(lineSlice)))
	mw.ExportBtn.SetEnabled(true)
	//正常还是断点 都要 读取源文件夹
	mw.ListFiles(sourceDirText)
	mw.fb.SetModel(mw.FindArr)
	//源文件夹文件数量
	mw.AllCountText.SetText(strconv.Itoa(len(mw.FindArr)))

	needMoveCount = 0

	mw.ExecBtn.SetText("匹配文件中...")

	findArrlen := len(mw.FindArr)
	i:=0
	for ;i<findArrlen;i++{
		for line,_:= range mw.lineMap{
			if mw.lineMap[line] !="" {
				continue
			}
			//已经存在就直接过 别浪费时间
			//判断
			fileName,_,_ := NewFile(mw.FindArr[i])
			if fileName == line{
				needMoveCount = needMoveCount +1
				mw.lineMap[line]=mw.FindArr[i]
				break
			}
		}
		mw.ExecBtn.SetText("匹配文件完成"+strconv.Itoa(mw.GetNum(findArrlen,i))+"%")
	}

	mw.ExecBtn.SetText("匹配完毕...")
	//需要拷贝的文件数量
	needMoveLen := strconv.Itoa(needMoveCount)
	//赋值
	mw.NeedCountText.SetText(needMoveLen)

	if !stopFlag{
		mw.TxtCountText.SetText(strconv.Itoa(len(mw.lineMap)))
	}

	mw.StopBtn.SetEnabled(true)

	stopFlag = false

	mw.ExecBtn.SetText("文件迁移中...")
	mw.TxtDealCountText.SetText("0")
	for line,needMoveFile:= range mw.lineMap{
		if stopFlag {
			return
		}
		if needMoveFile != ""{
			err := mw.CopeFile(needMoveFile)
			if err != nil{
				mw.Error("文件拷贝异常"+err.Error())
				return
			}
			count:=mw.DealCountText.Text()
			sum,_:= strconv.Atoi(count)
			sum++
			mw.DealCountText.SetText(strconv.Itoa(sum))
			delete(mw.lineMap,line)
		}

		txtDealCount:=mw.TxtDealCountText.Text()
		sum,_:= strconv.Atoi(txtDealCount)
		sum++
		mw.TxtDealCountText.SetText(strconv.Itoa(sum))
	}
	lineSlice = nil
	mw.ExecBtn.SetText("文件迁移完成，计算剩余配置数量中...")
	for k,_:= range mw.lineMap{
		lineSlice = append(lineSlice,k)
	}

	mw.lb.SetModel(lineSlice)
	mw.Success("处理完成")
}

//dos字符串转码
func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes,_=simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str= string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}

//拷贝文件
func(mw *MyWindow)CopeFile(needMoveFile string)(err error){

	_,fileNames,sourceDir :=NewFile(needMoveFile)

	copyFile := mw.TargetDirText.Text()+"\\"+fileNames

	exist,_:= mw.PathExists(copyFile)

	if mw.Baz == DUMP{
		if exist{
			return
		}
	}

	//多线程拷贝
	cmd := exec.Command("ROBOCOPY",sourceDir,mw.TargetDirText.Text(),fileNames,"/MT:128","/r:0")
	//剪切
	if mw.Baz == 2 {
		cmd = exec.Command("ROBOCOPY",sourceDir,mw.TargetDirText.Text(),fileNames,"/mov","/MT:128","/r:0")
	}
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	stdout, err := cmd.StdoutPipe()

	if  err != nil {//获取输出对象，可以从该对象中读取输出结果
		return err
	}

	defer stdout.Close()// 保证关闭输出流

	if err = cmd.Start(); err != nil {// 运行命令
		return err
	}

	if opBytes, err := ioutil.ReadAll(stdout);err != nil { // 读取输出结果   
		return err
	} else {
		cmdRe:=ConvertByte2String(opBytes,"GB18030")
		res := string(cmdRe)
		if strings.Contains(res,"错误"){
			errArr := strings.Split(res,"错误")
			return errors.New(errArr[1])
		}
	}
	return
}

//回收资源
func(mw *MyWindow) Close(){
	if !stopFlag{
		lineSlice = make([]string,0)
	}
	mw.FindArr = make([]string,0)
	mw.ExecBtn.SetText("执行")
	mw.ExecBtn.SetEnabled(true)
	mw.ExportBtn.SetEnabled(true)
}

//判断文件夹是否存在
func(mw *MyWindow) PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//文件解析
func NewFile(filePath string)(fileName string,fileNames string,Dir string){
	filePath2 := strings.Split(filePath,".")
	houzhui:=""
	if len(filePath2)>1{
		houzhui = filePath2[1]
	}
	filePath3 := strings.Split(filePath2[0],"\\")
	fileName = filePath3[len(filePath3)-1]
	fileNames = fileName+"."+houzhui
	Dir = strings.Split(filePath,fileNames)[0]
	return
}

//读取配置文件
func(mw *MyWindow) ReadLineFile(fileName string)(err error) {
	mw.ExecBtn.SetText("读取配置文件中...")
	mw.lineMap = make(map[string]string,0)
	if file, err := os.Open(fileName);err !=nil{
		return err
	}else {
		lineSlice = nil
		scanner := bufio.NewScanner(file)
		for scanner.Scan(){
			wg.Add(1)
			go func(line string){
				lock.Lock()
				mw.lineMap[line] = ""
				lineSlice = append(lineSlice,line)
				lock.Unlock()
				wg.Done()
			}(scanner.Text())
		}
		wg.Wait()
	}
	return err
}

//得到源文件夹 内的所有文件
func(mw *MyWindow) ListFiles(dirname string){
	mw.ExecBtn.SetText("遍历文件夹查找文件中...")
	files := make(chan string)
	fileWg.Add(1)
	go mw.WalkDir(dirname, files)
	go func(){
		defer close(files)
		fileWg.Wait()
	}()
	for file := range files {
		wg.Add(1)
		go func(file string) {
			lock.Lock()
			mw.FindArr = append(mw.FindArr,file)
			lock.Unlock()
			wg.Done()
		}(file)
	}
	wg.Wait()
}

//提示
func(mw *MyWindow) Error(msg string){walk.MsgBox(mw, "错误提示", msg, walk.MsgBoxIconError)}

//提示
func(mw *MyWindow) Success(msg string){walk.MsgBox(mw, "正确提示",msg , walk.MsgBoxIconInformation)}


func(mw *MyWindow) GetNum(fenmu int,funzi int)(i int){
	n2, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(funzi) / float64(fenmu)), 64)
	i = int(n2 * 100)
	return
}