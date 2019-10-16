package main

import (
	"bufio"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	LineArr []string
	NeedMoveArr []string
	ExportBtn *walk.PushButton
	lb *walk.ListBox
	ExportTxt  *walk.LineEdit
	DumpBtn *walk.RadioButton
	NoDumpBtn  *walk.RadioButton
	Dump int
	TxtDealCountText *walk.LineEdit
	StopBtn *walk.PushButton
}
var wg sync.WaitGroup
var fileWg sync.WaitGroup
var lineMap map[string]bool
var lineSlice []string
var err error
var lock sync.Mutex
var ch = make(chan struct{}, 255)
var stopFlag bool = false

const (
	txtFilePath int = 1
	exportFilePath int = 2
	sourceDirPath int = 1
	targetDirPath int = 2
)
func main() {
	mw := new(MyWindow)
	mw.Baz = 1
	mw.Dump = 1
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
				mw.InitTxt()
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
							Label{Text: "配置文件(也可将文件拖入窗体):"},
							LineEdit{AssignTo: &mw.TxtText},
							PushButton{
								AssignTo: &mw.ChooseTxtBtn,
								Text:     "选择配置文件",
							},
							Label{Text: "源文件夹(也可将文件拖入窗体):"},
							LineEdit{AssignTo: &mw.SourceDirText},
							PushButton{
								AssignTo: &mw.ChooseSourceBtn,
								Text:     "选择源文件夹",
							},
							Label{Text: "目标文件夹(也可将文件拖入窗体):"},
							LineEdit{AssignTo: &mw.TargetDirText},
							PushButton{
								AssignTo: &mw.ChooseTargetBtn,
								Text:     "选择目标文件夹",
							},
							Label{Text: "导出文件位置(也可将文件拖入窗体):"},
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
										Value: 1,
										AssignTo:&mw.DumpBtn,
									},
									RadioButton{
										Name:  "NoDump",
										Text:  "覆盖",
										Value: 2,
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
								Name:  "oneRB",
								Text:  "复制",
								Value: 1,
								AssignTo:&mw.CopeBtn,
							},
							RadioButton{
								Name:  "twoRB",
								Text:  "剪切",
								Value: 2,
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
				Layout: Grid{Columns: 1, Spacing: 80},
				Children: []Widget{
					ListBox{
						AssignTo: &mw.lb,
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
							Label{Text: "配置文件内的配置数量:"},
							LineEdit{AssignTo: &mw.TxtCountText, ReadOnly: true},
							Label{Text: "配置文件已完成数量:"},
							LineEdit{AssignTo: &mw.TxtDealCountText, ReadOnly: true},
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
	mw.ChooseTxtBtn.Clicked().Attach(func() {go mw.OpenFileActionTriggered(txtFilePath)})
	mw.ChooseExportBtn.Clicked().Attach(func() {go mw.OpenFileActionTriggered(exportFilePath)})
	mw.ChooseSourceBtn.Clicked().Attach(func() {go mw.OpenDirActionTriggered(sourceDirPath)})
	mw.ChooseTargetBtn.Clicked().Attach(func() {go mw.OpenDirActionTriggered(targetDirPath)})
	mw.Run()
}

//递归计算目录下所有文件
func(mw *MyWindow)WalkDir(path string, filePath chan <- string){
	defer fileWg.Done()
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
}

//暂停
func(mw *MyWindow)StopAction(){
	mw.Error("暂停成功")
	mw.StopBtn.SetEnabled(false)
	stopFlag= true
}

//让列表显示配置项
func(mw *MyWindow)InitTxt(){
	txtText := mw.TxtText.Text()
	err = mw.ReadLineFile(txtText)
	lineSlice = mw.LineArr
	mw.lb.SetModel(lineSlice)
	mw.TxtCountText.SetText(strconv.Itoa(len(lineSlice)))
	mw.ExportBtn.SetEnabled(true)
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
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, f := range lineSlice{
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
	mw.Success("文件导出完毕，文件地址为："+exportTxt)
}
//选择文件夹
func (mw *MyWindow)OpenDirActionTriggered(types int){
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
	if types == sourceDirPath{
		mw.SourceDirText.SetText(s)
	}else if types == targetDirPath{
		mw.TargetDirText.SetText(s)
	}
	mw.InitTxt()
}

//选择文件
func (mw *MyWindow)OpenFileActionTriggered(types int){
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
	if types == txtFilePath{
		mw.TxtText.SetText(s)
	}else if types == exportFilePath{
		mw.ExportTxt.SetText(s)
	}
	mw.InitTxt()
}

//执行
func (mw *MyWindow) ExecAction() {
	defer mw.Close()
	mw.NeedMoveArr = mw.NeedMoveArr[0:0]
	if !stopFlag{
		mw.TxtDealCountText.SetText("0")
		mw.DealCountText.SetText("0")
	}
	mw.ExecBtn.SetEnabled(false)
	mw.ExecBtn.SetText("执行中...")
	fmt.Println(mw.TxtText.Text())
	fmt.Println(mw.SourceDirText.Text())
	fmt.Println(mw.TargetDirText.Text())
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

	//读取文件
	err = mw.ReadLineFile(txtText)
	//读取源文件夹
	mw.ListFiles(sourceDirText)
	//源文件夹文件数量
	mw.AllCountText.SetText(strconv.Itoa(len(mw.FindArr)))

	for _,filePath := range mw.FindArr {
		wg.Add(1)
		go func(filePath string) {
			fileInfo := new(FileInfo)
			fileInfo.NewFile(filePath)
			fileName := fileInfo.FileName
			if mw.InLineArr(fileName){
				lock.Lock()
				mw.NeedMoveArr = append(mw.NeedMoveArr,filePath)
				lock.Unlock()
			}
			wg.Done()
		}(filePath)
	}

	wg.Wait()

	//需要拷贝的文件数量
	needMoveLen := len(mw.NeedMoveArr)
	needMoveLen2 := strconv.Itoa(needMoveLen)
	mw.NeedCountText.SetText(needMoveLen2)
	lineMap := make(map[string]bool,len(mw.LineArr))

	for _,line:= range mw.LineArr{
		wg.Add(1)
		go func(line string) {
			lock.Lock()
			lineMap[line]=true
			lock.Unlock()
			wg.Done()
		}(line)
	}

	wg.Wait()

	if !stopFlag{
		mw.TxtCountText.SetText(strconv.Itoa(len(mw.LineArr)))
	}
	mw.StopBtn.SetEnabled(true)
	stopFlag = false
	for _,line:= range mw.LineArr{
		flag := false
		for _,needMoveFile := range mw.NeedMoveArr{
			if stopFlag{
				mw.Success("暂停成功")
				return
			}
			fileInfo := new(FileInfo)
			fileInfo.NewFile(needMoveFile)
			if line == fileInfo.FileName{
				flag = true
				copyFile := mw.TargetDirText.Text()+"\\"+fileInfo.FileNames
				fileFlag,_ := mw.PathExists(copyFile)
				if fileFlag{
					if mw.Dump == 2{
						err := mw.CopeFile(needMoveFile,copyFile)
						if err!=nil{
							mw.Error("文件拷贝异常"+err.Error())
							return
						}
					}
				}else{
					err := mw.CopeFile(needMoveFile,copyFile)
					if err!=nil{
						mw.Error("文件拷贝异常"+err.Error())
						return
					}
				}
				if mw.Baz == 2{
					//删除文件
					os.Remove(needMoveFile)
				}
				count:=mw.DealCountText.Text()
				sum,_:= strconv.Atoi(count)
				sum++
				mw.DealCountText.SetText(strconv.Itoa(sum))
				break
			}
		}
		if flag {
			delete(lineMap, line)
			lineSlice = []string{}
			for k,_:= range lineMap{
				wg.Add(1)
				func(k string){
					lock.Lock()
					lineSlice = append(lineSlice,k)
					lock.Unlock()
					wg.Done()
				}(k)
			}
			wg.Wait()

			mw.lb.SetModel(lineSlice)
			txtDealCount:=mw.TxtDealCountText.Text()
			sum,_:= strconv.Atoi(txtDealCount)
			sum++
			mw.TxtDealCountText.SetText(strconv.Itoa(sum))
		}
	}
	mw.Success("处理完成")
}

//拷贝文件
func(mw *MyWindow)CopeFile(needMoveFile string,copyFile string)(err error){
	writeContent, err := ioutil.ReadFile(needMoveFile)
	if err!=nil{
		return err
	}
	err = ioutil.WriteFile(copyFile, []byte(writeContent), os.ModePerm)
	if nil != err {
		os.Remove(copyFile)
		return err
	}
	return
}

//回收资源
func(mw *MyWindow) Close(){
	if !stopFlag{
		mw.FindArr = mw.FindArr[0:0]
		mw.LineArr = mw.LineArr[0:0]
	}
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
func(mw *FileInfo)NewFile(filePath string){
	filePath2 := strings.Split(filePath,".")
	houzhui:=""
	if len(filePath2)>1{
		houzhui = filePath2[1]
	}

	filePath3 := strings.Split(filePath2[0],"\\")
	fileName := filePath3[len(filePath3)-1]
	filePath4 := strings.Split(filePath2[0],fileName)
	mw.Dir = filePath4[0]
	mw.Houzhui = houzhui
	mw.FileName = fileName
	mw.FileNames = fileName+"."+houzhui
}

//是否在配置文件中
func(mw *MyWindow) InLineArr(fileName string)bool{
	for _,v:= range mw.LineArr{
		if fileName == v{
			return true
		}
	}
	return false
}

//读取配置文件
func(mw *MyWindow) ReadLineFile(fileName string)(err error) {
	if stopFlag{
		mw.LineArr = lineSlice
	}else{
		if file, err := os.Open(fileName);err !=nil{
			return err
		}else {
			scanner := bufio.NewScanner(file)
			mw.LineArr = mw.LineArr[0:0]
			for scanner.Scan(){
				wg.Add(1)
				go func(line string){
					lock.Lock()
					mw.LineArr = append(mw.LineArr,string(line))
					lock.Unlock()
					wg.Done()
				}(scanner.Text())
			}
			wg.Wait()
		}
	}
	return err
}

//得到源文件夹 内的所有文件
func(mw *MyWindow) ListFiles(dirname string)  {

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

//得到源文件夹 内的所有文件
func(mw *MyWindow) listFiles(dirname string)  {
	fileInfos,_:=ioutil.ReadDir(dirname)
	for _,fi:=range fileInfos{
		filename := dirname+"\\"+fi.Name()   //拼写当前文件夹中所有的文件地址
		if fi.IsDir(){      //判断是否是文件夹 如果是继续调用把自己的地址作为参数继续调用
			mw.listFiles(filename)  //递归调用
		}else{
			mw.FindArr = append(mw.FindArr,filename)
			//fmt.Println(filename)    //打印文件地址
		}
	}
}

//提示
func(mw *MyWindow) Error(msg string){walk.MsgBox(mw, "错误提示", msg, walk.MsgBoxIconError)}

//提示
func(mw *MyWindow) Success(msg string){walk.MsgBox(mw, "正确提示",msg , walk.MsgBoxIconInformation)}
