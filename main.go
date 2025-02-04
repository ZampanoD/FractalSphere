package main

import (
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	rotationDir   float64 = 1.0
	scale         float64 = 50.0
	angle         float64 = 0
	isDragging    bool    = false
	offsetX       float64 = 0
	offsetY       float64 = 0
	windowX       int     = 0
	windowY       int     = 0
	sphereX       float64 = 300
	sphereY       float64 = 300
	initialWidth          = 600
	initialHeight         = 600
	lastMX        int     = 0
	lastMY        int     = 0
	lastWindowX   int     = 0
	lastWindowY   int     = 0
	vertexCache   sync.Map
)

type Point3D struct {
	X, Y, Z float64
}

type Face []int

type Game struct {
	offscreenImage *ebiten.Image
	needsRedraw    bool
	lastScale      float64
	lastAngle      float64
}

func NewGame() *Game {
	return &Game{
		offscreenImage: ebiten.NewImage(initialWidth, initialHeight),
		needsRedraw:    true,
		lastScale:      scale,
		lastAngle:      angle,
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		rotationDir = -1.0
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		rotationDir = 1.0
	}

	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		scale += scrollY * 10
		if scale < 10 {
			scale = 10
		}
		g.needsRedraw = true
	}

	if !isDragging {
		angle += 0.01 * rotationDir
		g.needsRedraw = true
	}

	if math.Abs(g.lastScale-scale) > 0.1 || math.Abs(g.lastAngle-angle) > 0.01 {
		g.needsRedraw = true
		g.lastScale = scale
		g.lastAngle = angle
	}

	mx, my := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if isCursorOverSphere(mx, my) {
			isDragging = true
			lastMX = mx
			lastMY = my
			lastWindowX, lastWindowY = ebiten.WindowPosition()
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && isDragging {
		if mx != lastMX || my != lastMY {
			deltaX := mx - lastMX
			deltaY := my - lastMY

			newX := lastWindowX + deltaX
			newY := lastWindowY + deltaY

			ebiten.SetWindowPosition(newX, newY)

			lastMX = mx
			lastMY = my
			lastWindowX = newX
			lastWindowY = newY
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		isDragging = false
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.needsRedraw {
		g.offscreenImage.Clear()
		g.drawSphere(g.offscreenImage)
		g.needsRedraw = false
	}
	screen.DrawImage(g.offscreenImage, nil)
}

func (g *Game) drawSphere(screen *ebiten.Image) {
	subdivisions := getLOD(scale)
	vertices, faces := getSphereData(subdivisions)

	timeVal := float64(ebiten.CurrentTPS()) / 60
	animatedVertices := animateVertices(vertices, timeVal)

	centerX, centerY := sphereX, sphereY
	points := make([][]Vec2, len(faces))
	for i, face := range faces {
		points[i] = make([]Vec2, len(face))
		for j, vIndex := range face {
			v := animatedVertices[vIndex]
			cosAngle := math.Cos(angle)
			sinAngle := math.Sin(angle)
			cosAngleY := math.Cos(angle * 0.8)
			sinAngleY := math.Sin(angle * 0.8)

			x := v.X*cosAngle - v.Z*sinAngle
			z := v.X*sinAngle + v.Z*cosAngle
			y := v.Y*cosAngleY - z*sinAngleY
			z = v.Y*sinAngleY + z*cosAngleY

			points[i][j] = Vec2{
				X: x*scale + centerX,
				Y: y*scale + centerY,
			}
		}
	}

	for i, facePoints := range points {
		t := (math.Sin(float64(i)*0.1+timeVal) + 1) / 2
		col := lerpColor(color.RGBA{65, 105, 225, 255}, color.RGBA{147, 112, 219, 255}, t)

		DrawOptimizedTriangle(screen, facePoints, col)
	}
}

func DrawOptimizedTriangle(screen *ebiten.Image, points []Vec2, col color.Color) {
	minX := int(math.Min(points[0].X, math.Min(points[1].X, points[2].X)))
	minY := int(math.Min(points[0].Y, math.Min(points[1].Y, points[2].Y)))
	maxX := int(math.Max(points[0].X, math.Max(points[1].X, points[2].X)))
	maxY := int(math.Max(points[0].Y, math.Max(points[1].Y, points[2].Y)))

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= initialWidth {
		maxX = initialWidth - 1
	}
	if maxY >= initialHeight {
		maxY = initialHeight - 1
	}

	for y := minY; y <= maxY; y++ {
		var intersections []float64
		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			if (points[i].Y <= float64(y) && points[j].Y > float64(y)) ||
				(points[j].Y <= float64(y) && points[i].Y > float64(y)) {
				x := points[i].X + (float64(y)-points[i].Y)*(points[j].X-points[i].X)/(points[j].Y-points[i].Y)
				intersections = append(intersections, x)
			}
		}
		if len(intersections) == 2 {
			if intersections[0] > intersections[1] {
				intersections[0], intersections[1] = intersections[1], intersections[0]
			}
			for x := int(intersections[0]); x <= int(intersections[1]); x++ {
				if x >= 0 && x < initialWidth {
					screen.Set(x, y, col)
				}
			}
		}
	}

	white := color.RGBA{255, 255, 255, 255}
	DrawLine(screen, points[0].X, points[0].Y, points[1].X, points[1].Y, white)
	DrawLine(screen, points[1].X, points[1].Y, points[2].X, points[2].Y, white)
	DrawLine(screen, points[2].X, points[2].Y, points[0].X, points[0].Y, white)
}

func getSphereData(subdivisions int) ([]Point3D, []Face) {
	key := subdivisions
	if cached, ok := vertexCache.Load(key); ok {
		if cachedData, ok := cached.(struct {
			vertices []Point3D
			faces    []Face
		}); ok {
			return cachedData.vertices, cachedData.faces
		}
	}

	vertices, faces := createSphere(subdivisions)
	vertexCache.Store(key, struct {
		vertices []Point3D
		faces    []Face
	}{vertices, faces})

	return vertices, faces
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return initialWidth, initialHeight
}

func getLOD(scale float64) int {
	if scale < 50 {
		return 1
	} else if scale < 100 {
		return 1
	} else if scale < 200 {
		return 2
	} else if scale < 400 {
		return 3
	}
	return 4
}

type Vec2 struct {
	X, Y float64
}

func createSphere(subdivisions int) ([]Point3D, []Face) {
	t := (1.0 + math.Sqrt(5.0)) / 2.0
	vertices := []Point3D{
		{-1, t, 0}, {1, t, 0}, {-1, -t, 0}, {1, -t, 0},
		{0, -1, t}, {0, 1, t}, {0, -1, -t}, {0, 1, -t},
		{t, 0, -1}, {t, 0, 1}, {-t, 0, -1}, {-t, 0, 1},
	}
	faces := []Face{
		{0, 11, 5}, {0, 5, 1}, {0, 1, 7}, {0, 7, 10}, {0, 10, 11},
		{1, 5, 9}, {5, 11, 4}, {11, 10, 2}, {10, 7, 6}, {7, 1, 8},
		{3, 9, 4}, {3, 4, 2}, {3, 2, 6}, {3, 6, 8}, {3, 8, 9},
		{4, 9, 5}, {2, 4, 11}, {6, 2, 10}, {8, 6, 7}, {9, 8, 1},
	}

	for i := 0; i < subdivisions; i++ {
		newFaces := []Face{}
		for _, face := range faces {
			mid01 := normalize(midpoint(vertices[face[0]], vertices[face[1]]))
			mid12 := normalize(midpoint(vertices[face[1]], vertices[face[2]]))
			mid20 := normalize(midpoint(vertices[face[2]], vertices[face[0]]))

			mid01Idx := len(vertices)
			vertices = append(vertices, mid01)
			mid12Idx := len(vertices)
			vertices = append(vertices, mid12)
			mid20Idx := len(vertices)
			vertices = append(vertices, mid20)

			newFaces = append(newFaces,
				Face{face[0], mid01Idx, mid20Idx},
				Face{face[1], mid12Idx, mid01Idx},
				Face{face[2], mid20Idx, mid12Idx},
				Face{mid01Idx, mid12Idx, mid20Idx},
			)
		}
		faces = newFaces
	}
	return vertices, faces
}

func normalize(p Point3D) Point3D {
	length := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
	return Point3D{p.X / length, p.Y / length, p.Z / length}
}

func midpoint(p1, p2 Point3D) Point3D {
	return Point3D{
		(p1.X + p2.X) / 2,
		(p1.Y + p2.Y) / 2,
		(p1.Z + p2.Z) / 2,
	}
}

func animateVertices(vertices []Point3D, time float64) []Point3D {
	animatedVertices := make([]Point3D, len(vertices))
	for i, vertex := range vertices {
		length := math.Sqrt(vertex.X*vertex.X + vertex.Y*vertex.Y + vertex.Z*vertex.Z)
		direction := Point3D{vertex.X / length, vertex.Y / length, vertex.Z / length}
		amplitude := 0.02
		frequency := 1.0
		phase := float64(i) * 0.1
		offset := amplitude * math.Sin(frequency*time+phase)
		animatedVertices[i] = Point3D{
			vertex.X + direction.X*offset,
			vertex.Y + direction.Y*offset,
			vertex.Z + direction.Z*offset,
		}
	}
	return animatedVertices
}

func lerpColor(c1, c2 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
		G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
		B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
		A: 255,
	}
}

func isPointInTriangle(px, py, ax, ay, bx, by, cx, cy float64) bool {
	area := 0.5 * (-by*cx + ay*(-bx+cx) + ax*(by-cy) + bx*cy)
	s := 1 / (2 * area) * (ay*cx - ax*cy + (cy-ay)*px + (ax-cx)*py)
	t := 1 / (2 * area) * (ax*by - ay*bx + (ay-by)*px + (bx-ax)*py)
	return s > 0 && t > 0 && (1-s-t) > 0
}

func isCursorOverSphere(mx, my int) bool {
	centerX, centerY := sphereX, sphereY
	radius := int(scale)
	dx, dy := mx-int(centerX), my-int(centerY)
	return dx*dx+dy*dy <= radius*radius
}

func DrawLine(screen *ebiten.Image, x1, y1, x2, y2 float64, col color.Color) {
	dx := x2 - x1
	dy := y2 - y1
	steps := math.Max(math.Abs(dx), math.Abs(dy))
	if steps == 0 {
		screen.Set(int(x1), int(y1), col)
		return
	}
	xIncrement := dx / steps
	yIncrement := dy / steps
	for i := 0; i <= int(steps); i++ {
		x := x1 + xIncrement*float64(i)
		y := y1 + yIncrement*float64(i)
		screen.Set(int(x), int(y), col)
	}
}

func main() {
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetRunnableOnUnfocused(true)

	game := NewGame()
	ebiten.SetWindowSize(initialWidth, initialHeight)
	ebiten.SetWindowTitle("Fractal Sphere with Infinite Zoom")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
