package main

import (
	"fmt"
	"nis3364/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	app      fyne.App
	window   fyne.Window
	client   *ChatClient
	chatData binding.String
	scroll   *container.Scroll
}

func NewGUI() *GUI {
	a := app.New()
	w := a.NewWindow("Chat Client")
	w.Resize(fyne.NewSize(600, 400))

	return &GUI{
		app:      a,
		window:   w,
		client:   NewChatClient(),
		chatData: binding.NewString(),
	}
}

func (g *GUI) Run() {
	g.showLogin()
	g.window.ShowAndRun()
}

func (g *GUI) showLogin() {
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Enter Username")

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Server IP (default: localhost)")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Server Port (default: 8080)")

	loading := widget.NewProgressBarInfinite()
	loading.Hide()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Server IP", Widget: ipEntry},
			{Text: "Server Port", Widget: portEntry},
			{Text: "Username", Widget: userEntry},
		},
		OnSubmit: func() {
			username := userEntry.Text
			if username == "" {
				dialog.ShowError(fmt.Errorf("username required"), g.window)
				return
			}

			ip := ipEntry.Text
			if ip == "" {
				ip = "localhost"
			}
			port := portEntry.Text
			if port == "" {
				port = "8080"
			}

			loading.Show()

			go func() {
				err := g.client.Connect(ip, port, username)
				if err != nil {
					g.window.Canvas().Refresh(loading)
					dialog.ShowError(err, g.window)
					loading.Hide()
					return
				}
				g.client.OnMessage = g.handleMessage
				g.client.OnError = func(err error) {
					dialog.ShowError(err, g.window)
					g.showLogin()
				}
				g.showChat()
			}()
		},
	}

	content := container.NewVBox(
		widget.NewLabelWithStyle("Welcome to Chat System", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		loading,
	)

	g.window.SetContent(container.NewCenter(container.NewGridWrap(fyne.NewSize(300, 300), content)))
}

func (g *GUI) showChat() {
	g.chatData.Set("Connected as " + g.client.username + "\n")
	chatLog := widget.NewLabelWithData(g.chatData)
	chatLog.Wrapping = fyne.TextWrapWord

	g.scroll = container.NewVScroll(chatLog)

	targetEntry := widget.NewEntry()
	targetEntry.SetPlaceHolder("Target User")

	msgEntry := widget.NewEntry()
	msgEntry.SetPlaceHolder("Type your message...")

	sendBtn := widget.NewButton("Send", func() {
		target := targetEntry.Text
		content := msgEntry.Text
		if target == "" {
			dialog.ShowError(fmt.Errorf("target user required"), g.window)
			return
		}
		if content == "" {
			dialog.ShowError(fmt.Errorf("message content required"), g.window)
			return
		}

		msg := utils.Message{
			Type:     "text",
			Content:  content,
			Sender:   g.client.username,
			Receiver: target,
		}
		g.client.Send(msg)
		msgEntry.SetText("")
	})

	broadcastBtn := widget.NewButton("Broadcast", func() {
		content := msgEntry.Text
		if content == "" {
			dialog.ShowError(fmt.Errorf("message content required"), g.window)
			return
		}
		msg := utils.Message{
			Type:    "broadcast",
			Content: content,
			Sender:  g.client.username,
		}
		g.client.Send(msg)
		msgEntry.SetText("")
	})

	askBtn := widget.NewButton("Ask For Current User", func() {
		msg := utils.Message{
			Type:   "ask",
			Sender: g.client.username,
		}
		g.client.Send(msg)
	})

	msgEntry.OnSubmitted = func(_ string) { sendBtn.OnTapped() }

	topBar := container.NewBorder(nil, nil, askBtn, sendBtn, targetEntry)
	botBar := container.NewBorder(nil, nil, nil, broadcastBtn, msgEntry)

	inputArea := container.NewVBox(topBar, botBar)

	content := container.NewBorder(nil, inputArea, nil, nil, g.scroll)
	g.window.SetContent(container.NewPadded(content))
}

func (g *GUI) handleMessage(msg utils.Message) {
	switch msg.Type {
	case "broadcast":
		g.appendChat(fmt.Sprintf("[Broadcast from %s]: %s", msg.Sender, msg.Content))
	case "text":
		g.appendChat(fmt.Sprintf("[%s]: %s", msg.Sender, msg.Content))
	case "error":
		dialog.ShowError(fmt.Errorf(msg.Content), g.window)
	case "accept":
		g.appendChat(fmt.Sprintf("[System]: %s", msg.Content))
	}
}

func (g *GUI) appendChat(msg string) {
	current, _ := g.chatData.Get()
	g.chatData.Set(current + msg + "\n")
	g.scroll.ScrollToBottom()
}
