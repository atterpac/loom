package ui

import (
	"fmt"

	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const splashLogo = `
__/\\\___________________/\\\\\____________/\\\\\_______/\\\\____________/\\\\_
 _\/\\\_________________/\\\///\\\________/\\\///\\\____\/\\\\\\________/\\\\\\_
  _\/\\\_______________/\\\/__\///\\\____/\\\/__\///\\\__\/\\\//\\\____/\\\//\\\_
   _\/\\\______________/\\\______\//\\\__/\\\______\//\\\_\/\\\\///\\\/\\\/_\/\\\_
    _\/\\\_____________\/\\\_______\/\\\_\/\\\_______\/\\\_\/\\\__\///\\\/___\/\\\_
     _\/\\\_____________\//\\\______/\\\__\//\\\______/\\\__\/\\\____\///_____\/\\\_
      _\/\\\______________\///\\\__/\\\_____\///\\\__/\\\____\/\\\_____________\/\\\_
       _\/\\\\\\\\\\\\\\\____\///\\\\\/________\///\\\\\/_____\/\\\_____________\/\\\_
        _\///////////////_______\/////____________\/////_______\///______________\///__
`

// SplashModal displays the splash screen for testing gradients and themes.
type SplashModal struct {
	*tview.Flex
	logoText      *tview.TextView
	statusText    *tview.TextView
	logoContainer *tview.Flex
	topSpacer     *tview.Box
	bottomSpacer  *tview.Box
	leftSpacer    *tview.Box
	rightSpacer   *tview.Box
	onClose       func()
	gradientType  int // 0=diagonal, 1=horizontal, 2=vertical, 3=reverse-diagonal
}

// NewSplashModal creates a new splash modal for testing.
func NewSplashModal() *SplashModal {
	sm := &SplashModal{
		Flex:         tview.NewFlex().SetDirection(tview.FlexRow),
		logoText:     tview.NewTextView(),
		statusText:   tview.NewTextView(),
		gradientType: 0,
	}
	sm.setup()
	return sm
}

// SetOnClose sets the callback when the modal is closed.
func (sm *SplashModal) SetOnClose(fn func()) *SplashModal {
	sm.onClose = fn
	return sm
}

func (sm *SplashModal) setup() {
	sm.logoText.SetDynamicColors(true)
	sm.logoText.SetTextAlign(tview.AlignLeft)
	sm.logoText.SetBackgroundColor(ColorBg())

	sm.statusText.SetDynamicColors(true)
	sm.statusText.SetTextAlign(tview.AlignCenter)
	sm.statusText.SetBackgroundColor(ColorBg())

	sm.updateLogo()
	sm.updateStatus()

	// Create spacer boxes with background color
	sm.topSpacer = tview.NewBox().SetBackgroundColor(ColorBg())
	sm.bottomSpacer = tview.NewBox().SetBackgroundColor(ColorBg())
	sm.leftSpacer = tview.NewBox().SetBackgroundColor(ColorBg())
	sm.rightSpacer = tview.NewBox().SetBackgroundColor(ColorBg())

	// Logo container centered
	sm.logoContainer = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(sm.leftSpacer, 0, 1, false).
		AddItem(sm.logoText, 90, 0, false).
		AddItem(sm.rightSpacer, 0, 1, false)
	sm.logoContainer.SetBackgroundColor(ColorBg())

	// Build layout
	sm.AddItem(sm.topSpacer, 0, 1, false)
	sm.AddItem(sm.logoContainer, 12, 0, false)
	sm.AddItem(sm.statusText, 3, 0, false)
	sm.AddItem(sm.bottomSpacer, 0, 1, false)
	sm.SetBackgroundColor(ColorBg())

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		sm.SetBackgroundColor(ColorBg())
		sm.logoText.SetBackgroundColor(ColorBg())
		sm.statusText.SetBackgroundColor(ColorBg())
		sm.logoContainer.SetBackgroundColor(ColorBg())
		sm.topSpacer.SetBackgroundColor(ColorBg())
		sm.bottomSpacer.SetBackgroundColor(ColorBg())
		sm.leftSpacer.SetBackgroundColor(ColorBg())
		sm.rightSpacer.SetBackgroundColor(ColorBg())
		sm.updateLogo()
		sm.updateStatus()
	})
}

func (sm *SplashModal) updateLogo() {
	colors := DefaultGradientTags()
	var gradientLogo string

	switch sm.gradientType {
	case 0:
		gradientLogo = ApplyDiagonalGradient(splashLogo, colors)
	case 1:
		gradientLogo = ApplyHorizontalGradient(splashLogo, colors)
	case 2:
		gradientLogo = ApplyVerticalGradient(splashLogo, colors)
	case 3:
		gradientLogo = ApplyReverseDiagonalGradient(splashLogo, colors)
	}

	sm.logoText.SetText(gradientLogo)
}

func (sm *SplashModal) updateStatus() {
	themeName := "unknown"
	if t := ActiveTheme(); t != nil {
		themeName = t.Key
	}

	gradientNames := []string{"diagonal", "horizontal", "vertical", "reverse-diagonal"}
	gradientName := gradientNames[sm.gradientType]

	sm.statusText.SetText(fmt.Sprintf(
		"[%s]Theme: [%s::b]%s[-:-:-]  [%s]Gradient: [%s::b]%s[-:-:-]\n[%s][T] Theme  [G] Gradient  [Esc] Close[-]",
		TagFgDim(), TagAccent(), themeName,
		TagFgDim(), TagAccent(), gradientName,
		TagFgDim(),
	))
}

func (sm *SplashModal) cycleGradient() {
	sm.gradientType = (sm.gradientType + 1) % 4
	sm.updateLogo()
	sm.updateStatus()
}

func (sm *SplashModal) showThemeSelector() {
	// Cycle through available themes
	themes := config.ThemeNames()
	currentTheme := ""
	if t := ActiveTheme(); t != nil {
		currentTheme = t.Key
	}

	// Find current index and go to next
	nextIndex := 0
	for i, name := range themes {
		if name == currentTheme {
			nextIndex = (i + 1) % len(themes)
			break
		}
	}

	_ = SetTheme(themes[nextIndex])
}

// InputHandler handles keyboard input.
func (sm *SplashModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return sm.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEscape:
			if sm.onClose != nil {
				sm.onClose()
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				if sm.onClose != nil {
					sm.onClose()
				}
			case 'T':
				sm.showThemeSelector()
			case 'G', 'g':
				sm.cycleGradient()
			}
		}
	})
}

// Draw fills the entire background before drawing children.
func (sm *SplashModal) Draw(screen tcell.Screen) {
	// Fill entire screen with background color
	width, height := screen.Size()
	bgStyle := tcell.StyleDefault.Background(ColorBg())
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw children on top
	sm.Flex.Draw(screen)
}

// Focus handles focus.
func (sm *SplashModal) Focus(delegate func(p tview.Primitive)) {
	sm.Flex.Focus(delegate)
}
