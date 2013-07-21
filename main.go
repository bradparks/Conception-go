package main

import (
	"fmt"
	. "gist.github.com/5286084.git"
	"log"
	"runtime"
	"time"

	_ "github.com/ftrvxmtrx/tga"
	"image"
	_ "image/png"
	"os"

	//"github.com/go-gl/gl"
	gl "github.com/chsc/gogl/gl21"
	glfw "github.com/go-gl/glfw3"

	"github.com/Jragonmiris/mathgl"

	"github.com/shurcooL/go-goon"

	. "gist.github.com/6003701.git"
)

var _ = UnderscoreSepToCamelCase
var _ = goon.Dump

var offX, offY gl.Float
var oFontBase gl.Uint

var widgets []Widgeter
var mousePointer *Pointer
var keyboardPointer *Pointer

func CheckGLError() {
	errorCode := gl.GetError()
	if 0 != errorCode {
		log.Panic("GL Error: ", errorCode)
	}
}

func Print(x, y float32, s string) {
	if 0 == len(s) {
		return
	}

	gl.Enable(gl.BLEND)
	gl.Enable(gl.TEXTURE_2D)

	gl.PushMatrix()
	gl.Translatef(gl.Float(x), gl.Float(y), 0)
	gl.Translatef(-4+0.25, -1, 0)
	gl.ListBase(oFontBase)
	gl.CallLists(gl.Sizei(len(s)), gl.UNSIGNED_BYTE, gl.Pointer(&[]byte(s)[0]))
	gl.PopMatrix()

	CheckGLError()

	gl.Disable(gl.BLEND)
	gl.Disable(gl.TEXTURE_2D)
}

func InitFont() {
	const fontWidth = 8

	LoadTexture("./data/Font2048.tga")

	oFontBase = gl.GenLists(256)

	for iLoop1 := 0; iLoop1 < 256; iLoop1++ {
		fCharX := gl.Float(iLoop1%16) / 16.0
		fCharY := gl.Float(iLoop1/16) / 16.0

		gl.NewList(oFontBase+gl.Uint(iLoop1), gl.COMPILE)
		const offset = gl.Float(0.004)
		//#if DECISION_RENDER_TEXT_VCENTERED_MID
		VerticalOffset := gl.Float(0.00125)
		if ('a' <= iLoop1 && iLoop1 <= 'z') || '_' == iLoop1 {
			VerticalOffset = gl.Float(-0.00225)
		}
		/*#else
					VerticalOffset := gl.Float(0.0)
		//#endif*/
		gl.Begin(gl.QUADS)
		gl.TexCoord2f(fCharX+offset, 1-(1-fCharY-0.0625+offset+VerticalOffset))
		gl.Vertex2i(0, 16)
		gl.TexCoord2f(fCharX+0.0625-offset, 1-(1-fCharY-0.0625+offset+VerticalOffset))
		gl.Vertex2i(16, 16)
		gl.TexCoord2f(fCharX+0.0625-offset, 1-(1-fCharY-offset+VerticalOffset))
		gl.Vertex2i(16, 0)
		gl.TexCoord2f(fCharX+offset, 1-(1-fCharY-offset+VerticalOffset))
		gl.Vertex2i(0, 0)
		gl.End()
		gl.Translated(fontWidth, 0.0, 0.0)
		gl.EndList()
	}

	CheckGLError()
}

func DeinitFont() {
	gl.DeleteLists(oFontBase, 256)
}

func LoadTexture(path string) {
	fmt.Printf("Trying to load texture %q.\n", path)

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	fmt.Printf("Loaded %vx%v texture.\n", bounds.Dx(), bounds.Dy())

	var pixPointer *uint8
	switch img := img.(type) {
	case *image.RGBA:
		pixPointer = &img.Pix[0]
	case *image.NRGBA:
		pixPointer = &img.Pix[0]
	default:
		panic("Unsupported type.")
	}

	var texture gl.Uint
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.GENERATE_MIPMAP, gl.TRUE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_LOD_BIAS, -0.65)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.Sizei(bounds.Dx()), gl.Sizei(bounds.Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Pointer(pixPointer))
	CheckGLError()
}

type Widgeter interface {
	Render()
	Hit(mathgl.Vec4f) []Widgeter
	HoverPointers() map[*Pointer]bool
}

type Widget struct {
	x, y          gl.Float
	dx, dy        gl.Float
	hoverPointers map[*Pointer]bool
}

func NewWidget(x, y, dx, dy gl.Float) Widget {
	return Widget{x, y, dx, dy, map[*Pointer]bool{}}
}

func (*Widget) Render() {}
func (w *Widget) Hit(ParentPosition mathgl.Vec4f) []Widgeter {
	Hit := (ParentPosition[0] >= float32(w.x) &&
		ParentPosition[1] >= float32(w.y) &&
		ParentPosition[0] < float32(w.x+w.dx) &&
		ParentPosition[1] < float32(w.y+w.dy))

	if (Hit) {
		return []Widgeter{w}
	} else {
		return nil
	}
}
func (w *Widget) HoverPointers() map[*Pointer]bool {
	return w.hoverPointers
}

type BoxWidget struct {
	Widget
}

func (w *BoxWidget) Render() {
	/*// HACK: Brute-force check the mouse pointer if it contains this widget
	isHit := false
	for _, hit := range mousePointer.Mapping {
		if &w.Widget == hit {
			isHit = true
			break
		}
	}*/
	isHit := len(w.HoverPointers()) > 0

	if !isHit {
		DrawBox(w.x, w.y, w.dx, w.dy)
	} else {
		DrawYBox(w.x, w.y, w.dx, w.dy)
	}
}

func DrawBox(x, y, dx, dy gl.Float) {
	gl.Color3f(0.3, 0.3, 0.3)
	gl.Rectf(x-1, y-1, x+dx+1, y+dy+1)
	gl.Color3f(1, 1, 1)
	gl.Rectf(x, y, x+dx, y+dy)
}
func DrawYBox(x, y, dx, dy gl.Float) {
	gl.Color3f(0.898, 0.765, 0.396)
	gl.Rectf(x-1, y-1, x+dx+1, y+dy+1)
	gl.Color3f(1, 1, 1)
	gl.Rectf(x, y, x+dx, y+dy)
}

type CompositeWidget struct {
	Widget
	Widgets []Widgeter
}

func (w *CompositeWidget) Render() {
	gl.PushMatrix()
	defer gl.PopMatrix()
	gl.Translatef(w.x, w.y, 0)

	for _, widget := range w.Widgets {
		widget.Render()
	}
}
func (w *CompositeWidget) Hit(ParentPosition mathgl.Vec4f) []Widgeter {
	LocalPosition := ParentPosition.Sub(mathgl.Vec4f{float32(w.x), float32(w.y)})

	hits := []Widgeter{}
	for _, widget := range w.Widgets {
		hits = append(hits, widget.Hit(LocalPosition)...)
	}

	return hits
}

type SomethingWidget struct {
	Widget
	Updated bool
}

func (w *SomethingWidget) Render() {
	gl.PushMatrix()
	defer gl.PopMatrix()
	gl.Translatef(w.x, w.y, 0)
	gl.Color3f(0.3, 0.3, 0.3)
	gl.Rectf(0-1, 0-1, 8*27+1, 16+1)
	if !w.Updated {
		gl.Color3f(1, 1, 1)
	} else {
		gl.Color3f(0.9, 1, 0.9)
	}
	gl.Rectf(0, 0, 8*27, 16)

	gl.Color3f(0, 0, 0)
	Print(0, 0, "Hello Conception 2 (Go)! :D Woot")
}

type Something2Widget struct {
	Widget
}

func (w *Something2Widget) Render() {
	gl.PushMatrix()
	defer gl.PopMatrix()
	gl.Translatef(w.x, w.y, 0)
	gl.Color3f(0, 0, 0)

	gl.Enable(gl.TEXTURE_2D)
	gl.Enable(gl.BLEND)
	gl.Begin(gl.TRIANGLE_FAN)
	{
		const size = 256
		gl.TexCoord2i(0, 0)
		gl.Vertex2i(0, 0)
		gl.TexCoord2i(1, 0)
		gl.Vertex2i(size, 0)
		gl.TexCoord2i(1, 1)
		gl.Vertex2i(size, size)
		gl.TexCoord2i(0, 1)
		gl.Vertex2i(0, size)
	}
	gl.End()
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}

type UnderscoreSepToCamelCaseWidget struct {
	Widget
	window *glfw.Window
}

func (w *UnderscoreSepToCamelCaseWidget) Render() {
	gl.PushMatrix()
	defer gl.PopMatrix()
	gl.Translatef(w.x, w.y, 0)

	//s := w.window.GetClipboardString()
	s := "get_clipboard_string"
	// E.g., get_clipboard_string -> GetClipboardString
	s += " -> " + UnderscoreSepToCamelCase(s)
	w.dx = gl.Float(8 * len(s))
	w.dy = 16

	gl.Color3f(0.3, 0.3, 0.3)
	gl.Rectf(0-1, 0-1, w.dx+1, w.dy+1)
	gl.Color3f(1, 1, 1)
	gl.Rectf(0, 0, w.dx, w.dy)

	gl.Color3f(0, 0, 0)
	Print(0, 0, s)
}

type SpinnerWidget struct {
	Widget
	Spinner uint32
}

func (w *SpinnerWidget) Render() {
	gl.PushMatrix()
	defer gl.PopMatrix()
	gl.Color3f(0, 0, 0)
	gl.Translatef(w.x, w.y, 0)
	//gl.Rotatef(float32(spinner), 0, 0, 1)
	gl.Rotatef(gl.Float(w.Spinner), 0, 0, 1)
	gl.Begin(gl.LINES)
	gl.Vertex2i(0, 0)
	gl.Vertex2i(0, 20)
	gl.End()
}

type VirtualCategory uint8

const (
	TYPING VirtualCategory = iota
	POINTING
)

type Pointer struct {
	VirtualCategory VirtualCategory
	Mapping         []Widgeter
}

type EventType uint8

const (
	BUTTON_EVENT EventType = iota
	CHARACTER_EVENT
	SLIDER_EVENT
	AXIS_EVENT
)

type InputEvent struct {
	Pointer    *Pointer
	EventTypes map[EventType]bool
	//DeviceId   uint8
	InputId uint16

	Buttons []bool
	Sliders []float64
	Axes    []float64
}

func ProcessInputEventQueue(inputEventQueue []InputEvent) []InputEvent {
	for len(inputEventQueue) > 0 {
		inputEvent := inputEventQueue[0]

		if inputEvent.Pointer.VirtualCategory == POINTING && inputEvent.InputId == 0 && inputEvent.EventTypes[AXIS_EVENT] {
			// TODO: Query pointer state rather than event properties
			Position := mathgl.Vec4f{float32(inputEvent.Axes[0]), float32(inputEvent.Axes[1])}

			// Clear previously hit widgets
			for _, widget := range inputEvent.Pointer.Mapping {
				delete(widget.HoverPointers(), inputEvent.Pointer)
			}
			inputEvent.Pointer.Mapping = []Widgeter{}

			// Recalculate currently hit widgets
			for _, widget := range widgets {
				inputEvent.Pointer.Mapping = append(inputEvent.Pointer.Mapping, widget.Hit(Position)...)
			}
			for _, widget := range inputEvent.Pointer.Mapping {
				widget.HoverPointers()[inputEvent.Pointer] = true
			}
		}

		if inputEvent.Pointer == mousePointer && inputEvent.InputId == 0 && inputEvent.EventTypes[BUTTON_EVENT] && inputEvent.Buttons[0] {
			fmt.Println("Left down!")
		} else if inputEvent.Pointer == mousePointer && inputEvent.InputId == 1 && inputEvent.EventTypes[BUTTON_EVENT] && inputEvent.Buttons[0] {
			fmt.Println("Right down!")
		}

		inputEventQueue = inputEventQueue[1:]
	}

	inputEventQueue = []InputEvent{}

	return inputEventQueue
}

func main() {
	runtime.LockOSThread()

	glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
		panic(fmt.Sprintf("glfw.ErrorCallback: %v: %v\n", err, desc))
	})

	if !glfw.Init() {
		panic("glfw.Init()")
	}
	defer glfw.Terminate()

	//glfw.OpenWindowHint(glfw.FsaaSamples, 32)
	window, err := glfw.CreateWindow(400, 400, "", nil, nil)
	CheckError(err)
	window.MakeContextCurrent()

	err = gl.Init()
	if nil != err {
		log.Print(err)
	}
	fmt.Println(gl.GoStringUb(gl.GetString(gl.VENDOR)), gl.GoStringUb(gl.GetString(gl.RENDERER)), gl.GoStringUb(gl.GetString(gl.VERSION)))

	//window.SetPosition(1600, 600)
	window.SetPosition(1275, 300)
	glfw.SwapInterval(1)

	InitFont()
	defer DeinitFont()

	redraw := true

	size := func(w *glfw.Window, width, height int) {
		fmt.Println("Framebuffer Size:", width, height)
		windowWidth, windowHeight := w.GetSize()
		fmt.Println("Window Size:", windowWidth, windowHeight)
		gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

		// Update the projection matrix
		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		gl.Ortho(0, gl.Double(windowWidth), gl.Double(windowHeight), 0, -1, 1)
		gl.MatrixMode(gl.MODELVIEW)

		redraw = true
	}
	window.SetFramebufferSizeCallback(size)
	width, height := window.GetFramebufferSize()
	size(window, width, height)

	box := &BoxWidget{NewWidget(50, 150, 16, 16)}
	widgets = append(widgets, box)
	spinner := SpinnerWidget{NewWidget(30, 30, 0, 0), 0}
	widgets = append(widgets, &spinner)
	something := SomethingWidget{NewWidget(50, 100, 0, 0), false}
	widgets = append(widgets, &something)
	widgets = append(widgets, &Something2Widget{NewWidget(50, 220, 0, 0)})
	widgets = append(widgets, &CompositeWidget{NewWidget(150, 150, 0, 0),
		[]Widgeter{
			&BoxWidget{NewWidget(0, 0, 16, 16)},
			&BoxWidget{NewWidget(16 + 2, 0, 16, 16)},
		},
	})
	widgets = append(widgets, &UnderscoreSepToCamelCaseWidget{NewWidget(50, 180, 0, 0), window})

	mousePointer = &Pointer{VirtualCategory: POINTING}
	keyboardPointer = &Pointer{VirtualCategory: TYPING}
	inputEventQueue := []InputEvent{}

	MousePos := func(w *glfw.Window, x, y float64) {
		redraw = true
		//fmt.Println("MousePos:", x, y)

		//(widgets[len(widgets)-1]).(*CompositeWidget).x = gl.Float(x)
		//(widgets[len(widgets)-1]).(*CompositeWidget).y = gl.Float(y)

		inputEvent := InputEvent{
			Pointer:    mousePointer,
			EventTypes: map[EventType]bool{AXIS_EVENT: true},
			//DeviceId:   0,
			InputId: 0,
			Buttons: nil,
			Sliders: nil,
			Axes:    []float64{x, y},
		}
		//fmt.Printf("%#v\n", inputEvent)
		inputEventQueue = append(inputEventQueue, inputEvent)
	}
	window.SetCursorPositionCallback(MousePos)

	window.SetScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
		offX += gl.Float(xoff * 10)
		offY += gl.Float(yoff * 10)
		redraw = true

		inputEvent := InputEvent{
			Pointer:    mousePointer,
			EventTypes: map[EventType]bool{SLIDER_EVENT: true},
			//DeviceId:   0,
			InputId: 2,
			Buttons: nil,
			Sliders: []float64{yoff, xoff},
			Axes:    nil,
		}
		//fmt.Printf("%#v\n", inputEvent)
		inputEventQueue = append(inputEventQueue, inputEvent)
	})

	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		inputEvent := InputEvent{
			Pointer:    mousePointer,
			EventTypes: map[EventType]bool{BUTTON_EVENT: true},
			//DeviceId:   0,
			InputId: uint16(button),
			Buttons: []bool{action != glfw.Release},
			Sliders: nil,
			Axes:    nil,
		}
		inputEventQueue = append(inputEventQueue, inputEvent)
	})

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		inputEvent := InputEvent{
			Pointer:    keyboardPointer,
			EventTypes: map[EventType]bool{BUTTON_EVENT: true},
			//DeviceId:   130,
			InputId: uint16(key),
			Buttons: []bool{action != glfw.Release},
			Sliders: nil,
			Axes:    nil,
		}
		fmt.Println(key, action, mods)
		inputEventQueue = append(inputEventQueue, inputEvent)
	})

	go func() {
		<-time.After(3 * time.Second)
		log.Println("trigger!")
		something.Updated = true
		redraw = true
	}()

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.8, 0.3, 0.01, 1)

	//last := window.GetClipboardString()

	for !window.ShouldClose() && glfw.Press != window.GetKey(glfw.KeyEscape) {
		//glfw.WaitEvents()
		glfw.PollEvents()

		/*now := window.GetClipboardString()
		if now != last {
			last = now
			redraw = true
			fmt.Println("GetClipboardString changed!")
		}*/

		// Input
		inputEventQueue = ProcessInputEventQueue(inputEventQueue)

		if redraw {
			redraw = false

			gl.Clear(gl.COLOR_BUFFER_BIT)
			gl.LoadIdentity()
			gl.Translatef(offX, offY, 0)

			for _, widget := range widgets {
				widget.Render()
			}

			window.SwapBuffers()
			spinner.Spinner++
			//log.Println("swapped buffers")
		} else {
			time.Sleep(time.Millisecond)
		}

		//runtime.Gosched()
	}
}