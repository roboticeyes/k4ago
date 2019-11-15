package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/roboticeyes/gorex/encoding/rex"
	"github.com/roboticeyes/k4ago"
	"golang.org/x/image/tiff"
)

const (
	min     = 500
	max     = 1000
	spacing = float32(0.005)
)

func countDevices() uint {
	deviceCount := k4ago.AvailableDevices()
	fmt.Printf("Found %d devices\n", deviceCount)
	return deviceCount
}

func main() {
	log.Println("k4ago")
	var err error

	if countDevices() == 0 {
		panic("No device found.")
	}
	d := k4ago.NewDevice(k4ago.Default)
	if err := d.Open(); err != nil {
		panic(err)
	}
	defer d.Close()

	// Get versions
	// versions, err := d.Versions()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println(versions)

	// Get serial number
	// serial, err := d.SerialNumber()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("Serial:", serial)

	// Start camera for capture mode
	err = d.Start()
	if err != nil {
		log.Println(err)
	}

	capture := k4ago.NewCapture(d)
	capture.SingleShot()

	colorImage := capture.ColorImage()
	colorFile, err := os.Create("color.png")
	if err != nil {
		log.Fatal(err)
	}
	defer colorFile.Close()
	png.Encode(colorFile, colorImage)

	// depthImage := capture.DepthImage()
	// depthFile, err := os.Create("depth.tif")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer depthFile.Close()
	// tiff.Encode(depthFile, depthImage, &tiff.Options{})

	depthTransformed := capture.DepthTransformed()
	depthTransformedFile, err := os.Create("depth_transformed.tif")
	if err != nil {
		log.Fatal(err)
	}
	defer depthTransformedFile.Close()
	tiff.Encode(depthTransformedFile, depthTransformed, &tiff.Options{})

	// Stop camera
	d.Stop()

	width := depthTransformed.Bounds().Dx()
	height := depthTransformed.Bounds().Dy()

	mesh := rex.Mesh{ID: 0}

	validMap := make(map[int]int)

	i := 0
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			depth := depthTransformed.Gray16At(x, y).Y
			if depth > min && depth < max {
				validMap[getHash(x, y, width)] = idx
				idx++
				xC := float32(x) * spacing
				yC := float32(y) * spacing
				zC := float32(depth) * 0.0035
				r := float32(colorImage.Pix[i]) / 255.0
				g := float32(colorImage.Pix[i+1]) / 255.0
				b := float32(colorImage.Pix[i+2]) / 255.0

				mesh.Coords = append(mesh.Coords, mgl32.Vec3{xC, -yC, zC})
				mesh.Colors = append(mesh.Colors, mgl32.Vec3{r, g, b})
			}
			i += 4
		}
	}

	// generate faces according to the validmap
	type vertex struct {
		valid bool
		idx   int
	}
	for y := 0; y < height-1; y++ {
		for x := 0; x < width-1; x++ {

			var v00, v10, v01, v11 vertex
			v00.idx, v00.valid = validMap[getHash(x, y, width)]
			v10.idx, v10.valid = validMap[getHash(x+1, y, width)]
			v01.idx, v01.valid = validMap[getHash(x, y+1, width)]
			v11.idx, v11.valid = validMap[getHash(x+1, y+1, width)]

			// triangle 1
			if v00.valid && v01.valid && v11.valid {
				mesh.Triangles = append(mesh.Triangles, rex.Triangle{
					V0: uint32(v00.idx),
					V1: uint32(v11.idx),
					V2: uint32(v01.idx),
				})
			}
			// triangle 2
			if v00.valid && v11.valid && v10.valid {
				mesh.Triangles = append(mesh.Triangles, rex.Triangle{
					V0: uint32(v00.idx),
					V1: uint32(v10.idx),
					V2: uint32(v11.idx),
				})
			}
		}
	}

	// assign material
	mat := rex.NewMaterial(1)
	mat.KdRgb = mgl32.Vec3{0.9, 0.7, 0.1}
	mesh.MaterialID = 1

	rexFile := rex.File{}
	rexFile.Meshes = append(rexFile.Meshes, mesh)
	rexFile.Materials = append(rexFile.Materials, mat)
	var rexBuf bytes.Buffer
	e := rex.NewEncoder(&rexBuf)
	err = e.Encode(rexFile)
	if err != nil {
		panic(err)
	}
	fout, _ := os.Create("face.rex")
	fout.Write(rexBuf.Bytes())
	defer fout.Close()
}

func getHash(x, y, width int) int {
	return y*width + x
}
