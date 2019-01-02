package main

import (
	"encoding/json"
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image/color"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func randomNiceColor() pixel.RGBA {
again:
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	len := math.Sqrt(r*r + g*g + b*b)
	if len == 0 {
		goto again
	}
	return pixel.RGB(r/len, g/len, b/len)
}

type region struct {
	rect  *pixel.Rect
	circle *[2]pixel.Vec
	color color.Color
	thickness float64
}

func (p *region) draw(imd *imdraw.IMDraw) {
	imd.Color = p.color
	if p.rect != nil {
		imd.Push(p.rect.Min, p.rect.Max)
		imd.Rectangle(p.thickness)
	}
	if p.circle != nil {
		imd.Push(p.circle[0])
		imd.Ellipse(p.circle[1], p.thickness)
	}
}

type Command struct {
	Name  string `json:"n"`
	Value string `json:"v"`
}

type Coords [4]int

type Element struct {
	C *Command `json:"c"`
	R *Coords  `json:"r"`
}

type Frame struct {
	Elements []Element `json:"e"`
}

type Shape int

var (
	Rectangle Shape = 0
	Circle Shape = 1
)

type Scene struct {
	BaseFrame *Frame `json:"b"`
	Frames []Frame `json:"f"`
}

type displayFrame struct {
	regions []region
}

func (f Frame) toDisplayFrame() displayFrame {
	df := displayFrame{}
	currentColor := pixel.RGB(float64(0),float64(0),float64(0))
	currentThickness := 0.0
	shapeMode := Rectangle
	for _, e := range f.Elements {
		if e.C != nil {
			if e.C.Name == "color" {
				rgb := strings.Split(e.C.Value, ",")
				if len(rgb) == 3 {
					r, _ := strconv.ParseFloat(rgb[0], 4)
					g, _ := strconv.ParseFloat(rgb[1], 4)
					b, _ := strconv.ParseFloat(rgb[2], 4)
					currentColor = pixel.RGB(r, g, b)
				} else if v, ok := colornames.Map[rgb[0]]; ok {
					// svg 1.1 lowercase name set supported
					currentColor = pixel.ToRGBA(v)
				}
			} else if e.C.Name == "thickness" {
				t, _ := strconv.ParseFloat(e.C.Value, 4)
				currentThickness = t
			} else if e.C.Name == "shape" {
				if e.C.Value == "rectangle" {
					shapeMode = Rectangle
				} else if e.C.Value == "circle" {
					shapeMode = Circle
				} else {
					fmt.Println("Unhandled shape mode passed")
				}
			}
		}
		if e.R != nil {
			if shapeMode == Rectangle {
				r := pixel.R(float64(e.R[0]), -float64(e.R[1]), float64(e.R[2]), -float64(e.R[3]))
				df.regions = append(df.regions, region{rect: &r, color:currentColor, thickness:currentThickness})
			} else if shapeMode == Circle {
				posEllipse := [2]pixel.Vec{{float64(e.R[0]), -float64(e.R[1])}, {float64(e.R[2]), -float64(e.R[3])}}
				df.regions = append(df.regions, region{circle:&posEllipse, color:currentColor, thickness:currentThickness})
			}
		}
	}
	return df
}


func run() {

	currentFrame := 0
	displayFrames := make([]displayFrame, 0)
	baseDisplayFrame := displayFrame{}
	cfg := pixelgl.WindowConfig{
		Title:  "Display Server",
		Bounds: pixel.R(0, -1024, 1024, 0),
		VSync:  true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		x1 := -1
		y1 := -1
		x2 := -1
		y2 := -1
		if val, ok := r.URL.Query()["x1"]; ok {
			v, _ := strconv.Atoi(val[0])
			x1 = v
		}
		if val, ok := r.URL.Query()["y1"]; ok {
			v, _ := strconv.Atoi(val[0])
			y1 = -v
		}
		if val, ok := r.URL.Query()["x2"]; ok {
			v, _ := strconv.Atoi(val[0])
			x2 = v
		}
		if val, ok := r.URL.Query()["y2"]; ok {
			v, _ := strconv.Atoi(val[0])
			y2 = -v
		}
		p := pixel.R(float64(x1), float64(y1), float64(x2), float64(y2))
		displayFrames[currentFrame].regions = append(displayFrames[currentFrame].regions, region{rect: &p})
		displayFrames[currentFrame].regions[len(displayFrames[currentFrame].regions)-1].color = randomNiceColor()
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/scene", func (w http.ResponseWriter, r *http.Request) {
		var m Scene
		if val, ok := r.URL.Query()["scene"]; ok {
			json.Unmarshal([]byte(val[0]), &m)
		} else if r.Method == "POST" {
			json.NewDecoder(r.Body).Decode(&m)
		}
		if m.BaseFrame != nil {
			baseDisplayFrame = m.BaseFrame.toDisplayFrame()
		} else {
			baseDisplayFrame = displayFrame{}
		}
		displayFrames = make([]displayFrame, len(m.Frames))
		for i, f := range m.Frames {
			displayFrames[i] = f.toDisplayFrame()
		}
		w.Write([]byte("ok"))
	})

	go func() {
		fmt.Println("serving on 8080")
		err := http.ListenAndServe("localhost:8080", nil)
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()
	var (
		defaultCamPos = pixel.V(float64(win.Bounds().Max.X/2),float64(win.Bounds().Min.Y/2))
		camPos       = defaultCamPos
		baseCamSpeed = 600.0
		camSpeed     = baseCamSpeed
		camZoom      = 1.0
		camZoomSpeed = 1.01
	)
	//panOffset := float64(100)
	txt := text.New(pixel.V(10, 20), text.NewAtlas(basicfont.Face7x13, text.ASCII))
	win.Clear(colornames.Skyblue)
	fps := time.Tick(time.Second / 30)
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.Pressed(pixelgl.KeyLeft) {
			currentFrame--
			if currentFrame < 0 {
				currentFrame = 0
			}
		}
		if win.Pressed(pixelgl.KeyRight) {
			currentFrame++
			if currentFrame > len(displayFrames)-1 {
				currentFrame = len(displayFrames)-1
			}
		}
		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		camSpeed = baseCamSpeed/camZoom
		win.SetMatrix(cam)
		if win.Pressed(pixelgl.KeyW) {
			camPos.Y += camSpeed * dt
			//win.SetBounds(pixel.R(win.Bounds().Min.X, win.Bounds().Min.Y + panOffset, win.Bounds().Max.X, win.Bounds().Max.Y + panOffset))
		}
		if win.Pressed(pixelgl.KeyS) {
			camPos.Y -= camSpeed * dt
			//win.SetBounds(pixel.R(win.Bounds().Min.X, win.Bounds().Min.Y - panOffset, win.Bounds().Max.X, win.Bounds().Max.Y - panOffset))
		}
		if win.Pressed(pixelgl.KeyA) {
			camPos.X -= camSpeed * dt
			//win.SetBounds(pixel.R(win.Bounds().Min.X - panOffset, win.Bounds().Min.Y , win.Bounds().Max.X - panOffset, win.Bounds().Max.Y))
		}
		if win.Pressed(pixelgl.KeyD) {
			camPos.X += camSpeed * dt
			//win.SetBounds(pixel.R(win.Bounds().Min.X + panOffset, win.Bounds().Min.Y, win.Bounds().Max.X + panOffset, win.Bounds().Max.Y))
		}
		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)
		// this will zero the screen for rendering (all our y values are negative to align with source coordinate system)
		// so to compensate, we move the screen 0,0 point to the upper left and make every supplied y negative
		if win.JustPressed(pixelgl.KeyZ) {
			//win.SetBounds(pixel.R(win.Bounds().Min.X, -1*(win.Bounds().Max.Y - win.Bounds().Min.Y), win.Bounds().Max.X, float64(0)))
			camPos = defaultCamPos
			camZoom = 1.0
		}
		txt.Clear()
		fmt.Fprintln(txt, "Frame", currentFrame)
		win.Clear(colornames.Skyblue)
		imd := imdraw.New(nil)
		txt.Draw(win, pixel.IM)

		//fmt.Println(win.GetPos()) position of the window on the screen
		for _, p := range baseDisplayFrame.regions {
			p.draw(imd)
		}
		if currentFrame > len(displayFrames) -1 {
			currentFrame = 0
		}
		if len(displayFrames) > 0 {
			for _, p := range displayFrames[currentFrame].regions {
				p.draw(imd)
			}
		}
		imd.Draw(win)
		win.Update()
		<-fps
	}
}

func main() {
	myScene := Scene{Frames:[]Frame{{Elements:[]Element{
		{C: &Command{Name: "color", Value:"0,0,0"}},
		{R: &Coords{1,2,3,4}},
	}}}}
	out, _ := json.Marshal(myScene)
	fmt.Println(string(out))
	/*

	two block with base frame test scene

	{
		"b": {"e":[{"r":[50,50,500,51]}]},
		"f":[
			{"e":[{"r":[100,200,300,400]}]},
			{"e":[{"c":{"n":"color","v":"255,0,0"}},{"r":[150,250,350,450]}]}
		]
	}

	 */
	pixelgl.Run(run)
}