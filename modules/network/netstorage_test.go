package network

import (
	"io/ioutil"
	"testing"
)

func TestReadUserArea(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/userarea.bin")
	if err != nil {
		t.Errorf("failed to open the test file: %v", err)
		return
	}

	area, err := ReadUserArea(data)
	if err != nil {
		t.Error(err)
		return
	}

	const expectedUnknown = 0
	if area.Unknown != expectedUnknown {
		t.Errorf("got Unknown %d expected %d", area.Unknown, expectedUnknown)
	}

	const expectedUnknownCount = 7
	if area.UnknownCount != expectedUnknownCount {
		t.Errorf("got UnknownCount %d expected %d", area.UnknownCount, expectedUnknownCount)
	}

	slot0 := area.Slots[0]
	expectedSlot0 := UserAreaSlot{
		Unknown: 1,
		Items: [10]UserAreaItem{
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
		},
		ItemsCount: 0,
		User:       0,
	}

	if slot0 != expectedSlot0 {
		t.Error("unexpected slot0")
	}

	slot5 := area.Slots[5]
	expectedSlot5 := UserAreaSlot{
		Unknown: 0,
		Items: [10]UserAreaItem{
			UserAreaItem(26),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
			UserAreaItem(0xFFFFFFFF),
		},
		ItemsCount: 1,
		User:       76561198028565520,
	}

	if slot5 != expectedSlot5 {
		t.Error("unexpected slot5")
	}

	return
}
