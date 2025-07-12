package colors

import "github.com/gdamore/tcell/v2"

// HLS syntax colors similar to hlsq
var (
	// Tag colors
	TagColor       = tcell.ColorYellow
	TagValueColor  = tcell.ColorWhite
	CommentColor   = tcell.ColorDarkGray
	
	// Manifest content colors
	URIColor       = tcell.ColorLightCyan
	BandwidthColor = tcell.ColorGreen
	ResolutionColor = tcell.ColorPurple
	CodecsColor    = tcell.ColorBlue
	DurationColor  = tcell.ColorGreen
	SequenceColor  = tcell.ColorAqua
	
	// Headers and UI colors
	HeaderColor    = tcell.ColorYellow
	HeaderBgColor  = tcell.ColorDarkBlue
	BorderColor    = tcell.ColorWhite
	SelectedColor  = tcell.ColorWhite
	ErrorColor     = tcell.ColorRed
	WarningColor   = tcell.ColorYellow
	SuccessColor   = tcell.ColorGreen
	
	// Encryption colors
	EncryptedColor = tcell.ColorRed
	UnencryptedColor = tcell.ColorDarkGray
)