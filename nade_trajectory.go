package main

import (
	"fmt"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/r3"
	"github.com/llgcode/draw2d/draw2dimg"
	ex "github.com/markus-wa/demoinfocs-golang/v3/examples"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/msg"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
)

type nadePath struct {
	wep  common.EquipmentType
	path []r3.Vector
	team common.Team
}

var (
	colorFireNade    color.Color = color.RGBA{R: 0xff, A: 0xff}
	colorInferno     color.Color = color.RGBA{R: 0xff, G: 0xa5, A: 0xff}
	colorInfernoHull color.Color = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
	colorHE          color.Color = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
	colorFlash       color.Color = color.RGBA{R: 0xff, A: 0xff}
	colorSmoke       color.Color = color.RGBA{R: 0xff, G: 0xbe, A: 0xff}
	colorDecoy       color.Color = color.RGBA{R: 0x95, G: 0x4b, A: 0xff}
)

var curMap ex.Map

func traceNadeTrajectory(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln("error while opening file ", err)
	}

	p := demoinfocs.NewParser(f)
	header, err := p.ParseHeader()
	if err != nil {
		log.Fatalln("error while parsing header ", err)
	}

	var (
		mapRadarImg image.Image
	)
	p.RegisterNetMessageHandler(func(msg *msg.CSVCMsg_ServerInfo) {
		curMap = ex.GetMapMetadata(header.MapName, msg.GetMapCrc())
		mapRadarImg = ex.GetMapRadar(header.MapName, msg.GetMapCrc())
	})

	nadeTrajectory := make(map[int64]*nadePath)

	p.RegisterEventHandler(func(e events.GrenadeProjectileDestroy) {
		id := e.Projectile.UniqueID()

		var team common.Team
		if e.Projectile.Thrower != nil {
			team = e.Projectile.Thrower.Team
		}

		if nadeTrajectory[id] == nil {
			nadeTrajectory[id] = &nadePath{
				wep:  e.Projectile.WeaponInstance.Type,
				path: nil,
				team: team,
			}
		}

		nadeTrajectory[id].path = e.Projectile.Trajectory
	})

	var infernos []*common.Inferno

	p.RegisterEventHandler(func(e events.InfernoExpired) {
		infernos = append(infernos, e.Inferno)
	})

	var nadeTrajectoryFirst5Rounds []*nadePath
	var infernosFirst5Rounds []*common.Inferno
	round := 0
	p.RegisterEventHandler(func(events.RoundEnd) {
		round++

		if round == 5 {
			for _, np := range nadeTrajectory {
				nadeTrajectoryFirst5Rounds = append(nadeTrajectoryFirst5Rounds, np)
			}
			nadeTrajectory = make(map[int64]*nadePath)

			infernosFirst5Rounds = make([]*common.Inferno, len(infernos))
			copy(infernosFirst5Rounds, infernos)
		}
	})

	err = p.ParseToEnd()
	if err != nil {
		log.Fatalln("error while parsing to end: ", err)
	}

	dest := image.NewRGBA(mapRadarImg.Bounds())
	draw.Draw(dest, dest.Bounds(), mapRadarImg, image.Point{}, draw.Src)

	gc := draw2dimg.NewGraphicContext(dest)

	drawInfernos(gc, infernosFirst5Rounds)
	drawTrajectories(gc, nadeTrajectoryFirst5Rounds)

	err = jpeg.Encode(os.Stdout, dest, &jpeg.Options{
		Quality: 90,
	})

	if err != nil {
		log.Fatalln("error while jpeg output")
	}
}

func drawInfernos(gc *draw2dimg.GraphicContext, infernos []*common.Inferno) {
	gc.SetFillColor(colorInferno)

	hulls := make([][]r2.Point, len(infernos))
	for i := range infernos {
		hulls[i] = infernos[i].Fires().ConvexHull2D()
	}

	for _, hull := range hulls {
		buildInfernoPath(gc, hull)
		gc.Fill()
	}

	gc.SetStrokeColor(colorInfernoHull)
	gc.SetLineWidth(1)

	for _, hull := range hulls {
		buildInfernoPath(gc, hull)
		gc.FillStroke()
	}
}

func buildInfernoPath(gc *draw2dimg.GraphicContext, vertices []r2.Point) {
	xOrigin, yOrigin := curMap.TranslateScale(vertices[0].X, vertices[0].Y)
	gc.MoveTo(xOrigin, yOrigin)

	for _, fire := range vertices[1:] {
		x, y := curMap.TranslateScale(fire.X, fire.Y)
		gc.LineTo(x, y)
	}
	gc.LineTo(xOrigin, yOrigin)
}

func drawTrajectories(gc *draw2dimg.GraphicContext, trajectories []*nadePath) {
	gc.SetLineWidth(1)
	gc.SetFillColor(color.RGBA{0, 0, 0, 0})

	for _, np := range trajectories {
		// set colors
		switch np.wep {
		case common.EqMolotov:
			fallthrough
		case common.EqIncendiary:
			gc.SetStrokeColor(colorFireNade)
		case common.EqHE:
			gc.SetStrokeColor(colorHE)
		case common.EqFlash:
			gc.SetStrokeColor(colorFlash)
		case common.EqSmoke:
			gc.SetStrokeColor(colorSmoke)
		case common.EqDecoy:
			gc.SetStrokeColor(colorDecoy)
		default:
			gc.SetStrokeColor(color.RGBA{})
			fmt.Println("unknown grenade type", np.wep)
		}

		x, y := curMap.TranslateScale(np.path[0].X, np.path[0].Y)
		gc.MoveTo(x, y)

		for _, pos := range np.path[1:] {
			x, y := curMap.TranslateScale(pos.X, pos.Y)
			gc.LineTo(x, y)
		}

		gc.FillStroke()
	}
}

func main() {
	traceNadeTrajectory("tmp/1662024657.dem")
}
