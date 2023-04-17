package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	hltv "github.com/Olament/HLTV-Go"
	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/msg"
	"os"
	"strconv"
)

type Output struct {
	Frame   int
	Events  interface{}
	Players [][]string
}

func csvToDemo(filename string) {
	f, err := os.Open(filename)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	if err != nil {
		panic(err)
	}

	p := dem.NewParser(f)

	var data []Output

	// parse frame by frame
	for ok := true; ok; ok, err = p.ParseNextFrame() {
		if err != nil {
			panic(err)
		}

		gs := p.GameState()
		frame := p.CurrentFrame()

		var players [][]string

		for _, player := range gs.Participants().Playing() {
			players = append(players, extractPlayerData(frame, player))
		}

		o := Output{
			Frame:   frame,
			Players: players,
		}
		data = append(data, o)
	}
	err = csvExport(data)
	if err != nil {
		panic(err)
	}
}

func csvExport(data []Output) error {
	file, err := os.OpenFile("results_new.csv", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	if err != nil {
		panic(err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"Frame", "Name", "SteamID", "Position_X", "Position_Y", "Position_Z", "LastAlivePosition_X",
		"LastAlivePosition_Y", "LastAlivePosition_Z", "Velocity_X", "Velocity_Y", "Velocity_Z", "ViewDirection_X",
		"ViewDirection_Z", "Health", "Armor", "Money", "CurrentEquipmentValue", "FreezeTimeEndEquipmentValue",
		"RoundStartEquipmentValue", "IsDucking", "HasDefuseKit", "HasHelmet", "Kills", "Deaths", "Assists", "Score",
		"MVPs", "MoneySpentTotal", "MoneySpentThisRound",
	}

	if err := writer.Write(header); err != nil {
		return err
	}
	// data
	for _, frameData := range data {
		for _, player := range frameData.Players {
			err := writer.Write(player)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func extractPlayerData(frame int, player *common.Player) []string {
	return []string{
		strconv.Itoa(frame),
		player.Name,
		strconv.FormatUint(player.SteamID64, 10),
		strconv.FormatFloat(player.Position().X, 'G', -1, 64),
		strconv.FormatFloat(player.Position().Y, 'G', -1, 64),
		strconv.FormatFloat(player.Position().Z, 'G', -1, 64),

		strconv.FormatFloat(player.LastAlivePosition.X, 'G', -1, 64),
		strconv.FormatFloat(player.LastAlivePosition.Y, 'G', -1, 64),
		strconv.FormatFloat(player.LastAlivePosition.Z, 'G', -1, 64),

		strconv.FormatFloat(player.Velocity().X, 'G', -1, 64),
		strconv.FormatFloat(player.Velocity().Y, 'G', -1, 64),
		strconv.FormatFloat(player.Velocity().Z, 'G', -1, 64),

		strconv.FormatFloat(float64(player.ViewDirectionX()), 'G', -1, 64),
		strconv.FormatFloat(float64(player.ViewDirectionY()), 'G', -1, 64),

		strconv.Itoa(player.Health()),
		strconv.Itoa(player.Armor()),
		strconv.Itoa(player.Money()),
		strconv.Itoa(player.EquipmentValueCurrent()),
		strconv.Itoa(player.EquipmentValueFreezeTimeEnd()),
		strconv.Itoa(player.EquipmentValueRoundStart()),
		strconv.FormatBool(player.IsDucking()),
		strconv.FormatBool(player.HasDefuseKit()),
		strconv.FormatBool(player.HasHelmet()),
		strconv.Itoa(player.Kills()),
		strconv.Itoa(player.Deaths()),
		strconv.Itoa(player.Assists()),
		strconv.Itoa(player.Score()),
		strconv.Itoa(player.MVPs()),
		strconv.Itoa(player.MoneySpentTotal()),
		strconv.Itoa(player.MoneySpentThisRound()),
	}
}

func eventsByFrame(filename string) {
	f, err := os.Open(filename)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	if err != nil {
		panic(err)
	}

	p := dem.NewParser(f)

	var eventsInCurrentFrame []interface{}

	// register handler on all events
	p.RegisterEventHandler(func(e interface{}) {
		eventsInCurrentFrame = append(eventsInCurrentFrame, e)
	})

	// parse frame by frame
	for ok := true; ok; ok, err = p.ParseNextFrame() {
		if err != nil {
			panic(err)
		}

		// iterate over events in frame
		for _, event := range eventsInCurrentFrame {
			handleEvent(event)
		}

		// reset event list
		eventsInCurrentFrame = eventsInCurrentFrame[:0]
	}
}

func handleEvent(event interface{}) {
	fmt.Println(event)
}

func testHLTV() {
	h := hltv.HLTV{
		Url:       "https://www.hltv.org",
		StaticURL: "",
	}
	events, _ := h.GetEvent(1)
	res, _ := json.MarshalIndent(events, "", " ")
	fmt.Println(string(res))
}

func demoHeader(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(fmt.Sprintf("error while closing file: %v", err))
		}
	}(f)

	p := dem.NewParser(f)

	header, err := p.ParseHeader()
	if err != nil {
		panic("error while parsing header")
	}
	fmt.Println(header.MapName)
	fmt.Println(header.ClientName)
	fmt.Println(header.ServerName)
	fmt.Println(header.GameDirectory)
	fmt.Println(header.Filestamp)
	fmt.Println(header.NetworkProtocol)
	fmt.Println(header.PlaybackFrames)
	fmt.Println(header.PlaybackTicks)
	fmt.Println(header.SignonLength)
	fmt.Println(header.Protocol)
	fmt.Println(header.FrameRate())
	fmt.Println(header.FrameTime())
}

func getMapCRCCode(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	p := dem.NewParser(f)

	p.RegisterNetMessageHandler(func(info *msg.CSVCMsg_ServerInfo) {
		fmt.Println("map_crc", info.MapCrc)
	})

	p.RegisterNetMessageHandler(func(info *msg.TournamentMatchSetup) {
		fmt.Println(info.GetEventId())
	})
	err = p.ParseToEnd()
	if err != nil {
		panic(err)
	}
}
func main() {
	//csvToDemo("tmp/1662024657.dem")
	eventsByFrame("tmp/1662024657.dem")
	//testHLTV()
	//demoHeader("tmp/1662024657.dem")
	//getMapCRCCode("tmp/1662024657.dem")

}
