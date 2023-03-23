package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

var gbDownTotal float64
var gbUpTotal float64
var mbUp float64
var mbDown float64
var mbpsDown []float64
var mbpsUp []float64
var data [][]float64

func updateTraffic() {
	go func() {
		rxBytes, err := ioutil.ReadFile("/sys/class/net/eth0/statistics/rx_bytes")
		if err != nil {
			log.Fatal(err)
		}
		txBytes, err := ioutil.ReadFile("/sys/class/net/eth0/statistics/tx_bytes")
		if err != nil {
			log.Fatal(err)
		}
		for {
			rxBytesNew, err := ioutil.ReadFile("/sys/class/net/eth0/statistics/rx_bytes")
			if err != nil {
				log.Fatal(err)
			}
			txBytesNew, err := ioutil.ReadFile("/sys/class/net/eth0/statistics/tx_bytes")
			if err != nil {
				log.Fatal(err)
			}
			// remove \n at end of string
			rxBytes = []byte(strings.TrimSuffix(string(rxBytes), "\n"))
			txBytes = []byte(strings.TrimSuffix(string(txBytes), "\n"))
			rxBytesNew = []byte(strings.TrimSuffix(string(rxBytesNew), "\n"))
			txBytesNew = []byte(strings.TrimSuffix(string(txBytesNew), "\n"))
			rxBytesInt, err := strconv.Atoi(string(rxBytes))
			if err != nil {
				log.Fatal(err)
			}
			txBytesInt, err := strconv.Atoi(string(txBytes))
			if err != nil {
				log.Fatal(err)
			}
			rxBytesNewInt, err := strconv.Atoi(string(rxBytesNew))
			if err != nil {
				log.Fatal(err)
			}
			txBytesNewInt, err := strconv.Atoi(string(txBytesNew))
			if err != nil {
				log.Fatal(err)
			}
			gbDownTotal = float64(rxBytesNewInt) / 1073741824
			gbUpTotal = float64(txBytesNewInt) / 1073741824
			mbDown = float64(rxBytesNewInt-rxBytesInt) / 1048576 * 2
			mbUp = float64(txBytesNewInt-txBytesInt) / 1048576 * 2
			if len(mbpsDown) >= 75 {
				mbpsDown = mbpsDown[1:]
			}
			if len(mbpsUp) >= 75 {
				mbpsUp = mbpsUp[1:]
			}
			mbpsDown = append(mbpsDown, float64(rxBytesNewInt-rxBytesInt)/2048)
			mbpsUp = append(mbpsUp, float64(txBytesNewInt-txBytesInt)/2048)
			rxBytes = rxBytesNew
			txBytes = txBytesNew
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func main() {
	updateTraffic()
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	sl1 := widgets.NewSparkline()
	sl1.Title = "Download"
	sl1.Data = mbpsDown
	sl1.LineColor = ui.ColorGreen

	sl2 := widgets.NewSparkline()
	sl2.Title = "Upload"
	sl2.Data = mbpsUp
	sl2.LineColor = ui.ColorRed
	sl2.TitleStyle.Fg = ui.ColorRed

	slg := widgets.NewSparklineGroup(sl1, sl2)
	slg.Title = "Network Usage"
	slg.SetRect(0, 0, 75, 18)

	p1 := widgets.NewParagraph()
	p1.Text = fmt.Sprintf("Download: %.2f MB/s", mbDown)
	p1.SetRect(0, 20, 50, 21)
	p1.Border = false

	p2 := widgets.NewParagraph()
	p2.Text = fmt.Sprintf("Upload: %.2f MB/s", mbUp)
	p2.SetRect(51, 20, 100, 21)
	p2.Border = false

	p3 := widgets.NewParagraph()
	p3.Text = fmt.Sprintf("Total Download: %.2f GB", gbDownTotal)
	p3.SetRect(0, 22, 50, 23)
	p3.Border = false

	p4 := widgets.NewParagraph()
	p4.Text = fmt.Sprintf("Total Upload: %.2f GB", gbUpTotal)
	p4.SetRect(0, 23, 50, 24)
	p4.Border = false

	poll := ui.PollEvents()
	go func() {
		for {
			sl1.Data = mbpsDown
			sl2.Data = mbpsUp
			if mbUp > 1024 {
				mbUp = mbUp / 1024
				p2.Text = fmt.Sprintf("Upload: %.2f GB/s", mbUp)
			} else {
				p2.Text = fmt.Sprintf("Upload: %.2f MB/s", mbUp)
			}
			if mbDown > 1024 {
				mbDown = mbDown / 1024
				p1.Text = fmt.Sprintf("Download: %.2f GB/s", mbDown)
			} else {
				p1.Text = fmt.Sprintf("Download: %.2f MB/s", mbDown)
			}
			p3.Text = fmt.Sprintf("Total Download: %.2f GB", gbDownTotal)
			p4.Text = fmt.Sprintf("Total Upload: %.2f GB", gbUpTotal)
			ui.Render(slg, p1, p2, p3, p4)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	for {
		e := <-poll
		if e.ID == "<C-c>" || e.ID == "q" {
			break
		}
	}
}
