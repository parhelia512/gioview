package widget

import (
	"image"
	"image/color"
	"strconv"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/theme"
)

type state uint8
type LabelAlignment uint8

const (
	inactive state = iota
	hovered
	activated
	focused
)

const (
	Top LabelAlignment = iota
	Left
	Right
)

type LabelOption struct {
	Alignment LabelAlignment
	// The max horizontal space the label occupies. Only valid
	// for Left or Right Alignment.
	MaxWidth unit.Dp

	// Space between label and input box
	Padding unit.Dp

	BoldText bool
}

// Another TextField implementation with the following features:
// * configurable padding and border radius
// * more compact design by put character counters inline.
// * subscribe ESC key events to defocus the text field.
// * configurable label alignment.
type TextField struct {
	// padding between the text and border.
	Padding unit.Dp
	// border radius of the input box.
	Radius     unit.Dp
	SingleLine bool
	// Text alignment in the box.
	Alignment text.Alignment
	// Label alignment option
	LabelOption LabelOption

	// Label text describes the TextField. Its location is dependent
	// on the LabelAlignment of LabelOption.
	Label string
	// Helper text to give additional context to a field.
	HelperText string
	// The maximum number of characters the text input will allow.
	// Zero means no limit.
	MaxChars int
	// Mask replaces the visual display of each rune in the contents with the given rune.
	Mask rune

	ErrorColor color.NRGBA

	// Leading appears before the content of the text input.
	Leading layout.Widget
	// Trailing appears after the content of the text input.
	Trailing layout.Widget
	// Prefer not to export editor to users.
	editor widget.Editor

	// click detects when the mouse pointer clicks or hovers
	// within the textfield.
	click gesture.Click
	// cached text from the editor, updated when editor content is changed
	text   string
	border border

	state     state
	changed   bool
	submitted bool
	errorMsg  string
}

type border struct {
	Thickness unit.Dp
	Color     color.NRGBA
}

func (in *TextField) init() {
	if in.Radius == 0 {
		in.Radius = unit.Dp(4)
	}
	if in.Padding == 0 {
		in.Padding = unit.Dp(12)
	}
	if in.MaxChars < 0 {
		in.MaxChars = 0
	}

	if in.LabelOption.Padding < 0 {
		in.LabelOption.Padding = unit.Dp(12)
	}

	if in.SingleLine != in.editor.SingleLine {
		in.editor.SingleLine = in.SingleLine
	}

	if in.Alignment != in.editor.Alignment {
		in.editor.Alignment = in.Alignment
	}

	if in.MaxChars != in.editor.MaxLen {
		in.editor.MaxLen = in.MaxChars
	}
	if in.Mask != in.editor.Mask {
		in.editor.Mask = in.Mask
	}

	if in.ErrorColor == (color.NRGBA{}) {
		in.ErrorColor = color.NRGBA{R: 200, A: 255}
	}

	// Enable submit if editor is single line.
	if in.SingleLine {
		in.editor.Submit = true
	}
}

func (in *TextField) update(gtx layout.Context, th *theme.Theme) {
	disabled := gtx.Source == (input.Source{})
	for {
		ev, ok := in.click.Update(gtx.Source)
		if !ok {
			break
		}
		switch ev.Kind {
		case gesture.KindPress:
			gtx.Execute(key.FocusCmd{Tag: &in.editor})
		}
	}

	in.state = inactive
	if in.click.Hovered() && !disabled {
		in.state = hovered
	}
	if in.editor.Len() > 0 {
		in.state = activated
	}
	if gtx.Source.Focused(&in.editor) && !disabled {
		in.state = focused
	}

	switch in.state {
	case inactive:
		in.border = border{
			Thickness: unit.Dp(0.5),
			Color:     misc.WithAlpha(th.Fg, 128),
		}
	case hovered:
		in.border = border{
			Thickness: unit.Dp(0.5),
			Color:     misc.WithAlpha(th.Fg, 221),
		}
	case focused:
		in.border = border{
			Thickness: unit.Dp(2),
			Color:     th.ContrastBg,
		}
	case activated:
		in.border = border{
			Thickness: unit.Dp(0.5),
			Color:     misc.WithAlpha(th.Fg, 221),
		}
	}

	// Update text if editor content is changed or editor is submitted.
	for {
		event, ok := in.editor.Update(gtx)
		if !ok {
			break
		}

		switch event.(type) {
		case widget.SubmitEvent:
			in.submitted = true
		case widget.ChangeEvent:
		default:
			continue
		}
		in.changed = true
		in.text = in.editor.Text()
		break
	}

	// Catch the ESC key being pressed to release the widget's focus.
	for {
		evt, ok := gtx.Event(key.Filter{Focus: &in.editor, Name: key.NameEscape})
		if !ok {
			break
		}

		switch ev := evt.(type) {
		case key.Event:
			if ev.Name == key.NameEscape {
				gtx.Execute(key.FocusCmd{})
				break
			}
		}
	}
}

func (in *TextField) Layout(gtx layout.Context, th *theme.Theme, hint string) layout.Dimensions {
	in.init()
	in.update(gtx, th)
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	gtx.Constraints.Min.Y = 0

	macro := op.Record(gtx.Ops)
	dims := in.layout(gtx, th, hint)
	call := macro.Stop()

	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	in.click.Add(gtx.Ops)
	event.Op(gtx.Ops, &in.editor)
	call.Add(gtx.Ops)
	return dims
}

func (in *TextField) layout(gtx layout.Context, th *theme.Theme, hint string) layout.Dimensions {
	switch in.LabelOption.Alignment {
	case Left, Right:
		return in.layoutHorizontal(gtx, th, hint)
	case Top:
		return in.layoutVertical(gtx, th, hint)
	}

	return D{}
}

func (in *TextField) layoutHorizontal(gtx layout.Context, th *theme.Theme, hint string) layout.Dimensions {
	labelPadding := in.LabelOption.Padding
	if in.Label == "" {
		labelPadding = 0
	}

	var direction layout.Direction
	switch in.LabelOption.Alignment {
	case Left:
		direction = layout.W
	case Right:
		direction = layout.E
	default:
		direction = layout.W
	}

	layoutLabel := func(gtx C) D {
		if in.Label == "" {
			return D{}
		}

		if in.LabelOption.MaxWidth > 0 {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(in.LabelOption.MaxWidth))
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}

		return direction.Layout(gtx, func(gtx C) D {
			label := material.Caption(th.Theme, in.Label)
			if in.LabelOption.BoldText {
				label.Font.Weight = font.Bold
			}
			return label.Layout(gtx)
		})
	}

	measureLabel := func(gtx C) D {
		if in.Label == "" {
			return D{}
		}

		macro := op.Record(gtx.Ops)
		dims := layoutLabel(gtx)
		_ = macro.Stop()
		return dims
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layoutLabel(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: labelPadding}.Layout),
				layout.Rigid(func(gtx C) D {
					return in.layoutMain(gtx, th, hint)
				}),
			)

		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					dims := measureLabel(gtx)
					return D{Size: image.Pt(dims.Size.X, 0)}
				}),
				layout.Rigid(layout.Spacer{Width: labelPadding}.Layout),
				layout.Rigid(func(gtx C) D {
					return in.layoutHelper(gtx, th)
				}),
			)
		}),
	)

}

func (in *TextField) layoutVertical(gtx layout.Context, th *theme.Theme, hint string) layout.Dimensions {
	labelPadding := in.LabelOption.Padding
	if in.Label == "" {
		labelPadding = 0
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if in.Label == "" || in.LabelOption.Alignment != Top {
				return layout.Dimensions{}
			}

			label := material.Caption(th.Theme, in.Label)
			if in.LabelOption.BoldText {
				label.Font.Weight = font.Bold
			}
			return label.Layout(gtx)

		}),

		layout.Rigid(layout.Spacer{Height: labelPadding}.Layout),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return in.layoutMain(gtx, th, hint)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return in.layoutHelper(gtx, th)
		}),
	)

}

func (in *TextField) layoutMain(gtx layout.Context, th *theme.Theme, hint string) layout.Dimensions {
	border := widget.Border{
		Color:        in.border.Color,
		Width:        in.border.Thickness,
		CornerRadius: in.Radius,
	}

	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(in.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if in.Leading != nil {
						return layout.Inset{Right: in.Padding}.Layout(gtx, in.Leading)
					}
					return layout.Dimensions{}
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(th.Theme, &in.editor, hint)
					ed.HintColor = misc.WithAlpha(th.Fg, 0x60)
					return ed.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if in.MaxChars <= 0 {
						return layout.Dimensions{}
					}
					return layout.Inset{
						Left: in.Padding,
					}.Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							count := material.Label(
								th.Theme,
								th.TextSize*0.9,
								strconv.Itoa(in.editor.Len())+"/"+strconv.Itoa(int(in.MaxChars)),
							)
							count.Color = misc.WithAlpha(th.Fg, 0xb6)
							return count.Layout(gtx)
						},
					)
				}),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if in.Trailing != nil && in.state >= activated {
						return layout.Inset{Left: in.Padding}.Layout(gtx, in.Trailing)
					}
					return layout.Dimensions{}
				}),
			)
		})
	})

}

func (in *TextField) layoutHelper(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	if in.HelperText == "" && in.errorMsg == "" {
		return layout.Dimensions{}
	}

	helper := in.errorMsg
	helperColor := in.ErrorColor
	if helper == "" {
		helper = in.HelperText
		helperColor = misc.WithAlpha(th.Fg, 0xb6)
	}

	return layout.Inset{
		Top:  unit.Dp(4),
		Left: in.Padding,
	}.Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			helper := material.Label(th.Theme, th.TextSize*0.9, helper)
			helper.Color = helperColor
			return helper.Layout(gtx)
		},
	)
}

func (in *TextField) State() *widget.Editor {
	return &in.editor
}

func (in *TextField) Focused(gtx layout.Context) bool {
	return gtx.Focused(in.editor)
}

func (in *TextField) SetFocus(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &in.editor})
}

// Clear clears the input text.
func (in *TextField) Clear() {
	in.changed = true
	in.text = ""
	in.editor.SetText("")
	in.errorMsg = ""
}

// Text returns the current input text.
func (in *TextField) Text() string {
	return in.text
}

func (in *TextField) SetText(text string) {
	in.changed = true
	in.text = text
	in.editor.SetText(text)
	in.errorMsg = ""
}

func (in *TextField) SetError(err string) {
	in.errorMsg = err
}

func (in *TextField) ClearError() {
	in.errorMsg = ""
}

// Changed returns whether or not the text input has changed since last call.
func (in *TextField) Changed() bool {
	changed := in.changed
	in.changed = false
	return changed
}

func (in *TextField) Submitted() bool {
	// submit is disabled.
	if !in.editor.Submit {
		return false
	}
	submitted := in.submitted
	in.submitted = false
	in.changed = false
	return submitted
}
