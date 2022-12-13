package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/amdf/ixxatvci3"
	"github.com/amdf/ixxatvci3/candev"
)

var canOk = make(chan int)
var can *candev.Device
var b candev.Builder

// var CODE uint64 = 0xFEEDC000000000DE

var bOkCAN bool
var bConnected bool

var labelConnect = widget.NewLabel("")
var labelText = widget.NewLabel("")
var btnPass *widget.Button

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Программа доступа к настройкам ИПТМ-395.0")
	w.Resize(fyne.NewSize(400, 400))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	btnPass = widget.NewButton("Активация", func() {
		msg := candev.Message{ID: KEY_ID, Len: 8, Data: [8]byte{
			uint8(CODE >> 56),
			uint8(CODE >> 48),
			uint8(CODE >> 40),
			uint8(CODE >> 32),
			uint8(CODE >> 24),
			uint8(CODE >> 16),
			uint8(CODE >> 8),
			uint8(CODE),
		}}
		can.Send(msg)
		btnPass.Disable()
	})

	// labelNull := widget.NewLabel("")

	btn := container.New(
		layout.NewMaxLayout(),
		canvas.NewRectangle(color.RGBA{255, 0, 0, 250}),
		btnPass,
	)
	btn.Resize(fyne.NewSize(200, 100)) // указываем размер кнопки
	btn.Move(fyne.NewPos(100, 200))    // указываем позицию положение кнопки

	topLabel := container.NewVBox(
		labelConnect,
		labelText,
	)

	content := container.NewWithoutLayout(topLabel, btn) // создаем специальный контейнер в котором будут размещаться элементы (положение)

	w.SetContent(content)

	go connectCAN()
	go processCAN()
	go processScreen()

	defer func() {
		bOkCAN = false
		resetInfo()
		can.Stop()
	}()

	w.ShowAndRun()
}

func resetInfo() {
	bConnected = false
}

func processScreen() {
	sec := time.NewTicker(200 * time.Millisecond)
	for range sec.C {

		stringConnected := ""
		stringText := ""

		if bOkCAN {
			if bConnected {
				stringConnected = "Соединено с ИПТМ-395"
				btnPass.Enable()

			} else {

				stringConnected = "Ожидание соединения с ИПТМ-395..."
				stringText = "Запустите прибор в режиме настройки"
				btnPass.Disable()
			}
		} else {
			stringConnected = "Не обнаружен адаптер USB-to-CAN"
			stringText = "Подключите адаптер, перезапустите программу"
			btnPass.Disable()
		}

		labelConnect.SetText(stringConnected)
		labelText.SetText(stringText)
	}
}

// Определяем наличие CAN адаптера
func connectCAN() {
	var err error
	can, err = b.Speed(ixxatvci3.Bitrate25kbps).Get()
	for {
		if err == nil {
			can.Run()
			canOk <- 1
			bOkCAN = true
			time.Sleep(500 * time.Millisecond)
		} else {
			bOkCAN = false
			// can, err = b.Speed(ixxatvci3.Bitrate25kbps).Get()
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func processCAN() {
	<-canOk
	ch, _ := can.GetMsgChannelCopy()

	for msg := range ch {

		// Проверяем наличие связи с ИПТМ по выдаваемым им сообщениям
		if (msg.ID == IPTM_ID) && (msg.Len == IPTM_LEN) && (msg.Data[0] == 0x00) {
			bConnected = true
			time.Sleep(2 * time.Second)
		} else {
			resetInfo()
		}
	}
}
