package main

import (
	"github.com/golang/geo/r2"
	ex "github.com/markus-wa/demoinfocs-golang/v3/examples"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/msg"
	"github.com/markus-wa/go-heatmap/v2"
	"github.com/markus-wa/go-heatmap/v2/schemes"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
)

const (
	killerDotSize     = 15
	killerOpacity     = 128
	killerJpegQuality = 90
)

func killerHeatMap(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln("error while Opening file")
	}

	p := demoinfocs.NewParser(f)

	header, err := p.ParseHeader()
	if err != nil {
		log.Fatalln("parsing header file failed", err)
	}

	var (
		mapMetadata ex.Map
		mapRadarImg image.Image
	)

	p.RegisterNetMessageHandler(func(msg *msg.CSVCMsg_ServerInfo) {
		mapMetadata = ex.GetMapMetadata(header.MapName, msg.GetMapCrc())
		mapRadarImg = ex.GetMapRadar(header.MapName, msg.GetMapCrc())
	})

	var points []r2.Point

	p.RegisterEventHandler(func(e events.ItemDrop) {
		x, y := mapMetadata.TranslateScale(e.Player.Position().X, e.Player.Position().Y)
		points = append(points, r2.Point{
			X: x,
			Y: y,
		})
	})

	err = p.ParseToEnd()
	if err != nil {
		log.Fatalln("error while parsing to end", err)
	}
	r2Bounds := r2.RectFromPoints(points...)
	padding := float64(killerDotSize) / 2.0
	bounds := image.Rectangle{
		Min: image.Point{X: int(r2Bounds.X.Lo - padding), Y: int(r2Bounds.Y.Lo - padding)},
		Max: image.Point{X: int(r2Bounds.X.Lo + padding), Y: int(r2Bounds.Y.Hi + padding)},
	}

	data := make([]heatmap.DataPoint, 0, len(points))

	for _, p := range points[1:] {
		data = append(data, heatmap.P(p.X, p.Y*-1))
	}

	img := image.NewRGBA(mapRadarImg.Bounds())
	draw.Draw(img, mapRadarImg.Bounds(), mapRadarImg, image.Point{}, draw.Over)

	imgHeatmap := heatmap.Heatmap(image.Rect(0, 0, bounds.Dx(), bounds.Dy()), data, killerDotSize, killerOpacity, schemes.AlphaFire)
	draw.Draw(img, bounds, imgHeatmap, image.Point{}, draw.Over)

	err = jpeg.Encode(os.Stdout, img, &jpeg.Options{Quality: killerJpegQuality})
	if err != nil {
		log.Fatalln("error while writing to jpeg ", err)
	}
}

func main() {
	killerHeatMap("tmp/1662024657.dem")
}
