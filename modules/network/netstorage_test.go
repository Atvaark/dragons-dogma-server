package network

import (
	"io/ioutil"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	testData, err := ioutil.ReadFile("testdata/userarea.bin")
	if err != nil {
		t.Errorf("failed to open the test file: %v", err)
		return
	}

	testArea, err := ReadUserArea(testData)
	if err != nil {
		t.Error(err)
		return
	}

	data, err := WriteUserArea(testArea)
	if err != nil {
		t.Error(err)
		return
	}

	area, err := ReadUserArea(data)
	if err != nil {
		t.Error(err)
		return
	}

	compareArea(t, area, testArea)
}

func TestWriteUserArea(t *testing.T) {
	var area1 UserArea
	area1.Unknown = 1
	area1.UnknownCount = 2
	slot1 := &area1.Slots[1]
	slot1.Unknown = 0
	for i := 0; i < len(slot1.Items); i++ {
		slot1.Items[i] = 0xFFFFFFFF
	}
	slot1.Items[0] = 10
	slot1.ItemsCount = 1
	slot1.User = 123456789123456789

	data, err := WriteUserArea(&area1)
	if err != nil {
		t.Errorf("failed to write user area: %v", err)
		return
	}

	readArea, err := ReadUserArea(data)
	if err != nil {
		t.Errorf("failed to read user area: %v", err)
		return
	}

	compareArea(t, readArea, &area1)
}

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

	compareSlot(t, &slot0, &expectedSlot0, 0)

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

	compareSlot(t, &slot5, &expectedSlot5, 05)

	return
}

func compareArea(t *testing.T, actual, expected *UserArea) {
	if actual.Unknown != expected.Unknown {
		t.Errorf("got Unknown %v expected %v", actual.Unknown, expected.Unknown)
	}

	if actual.UnknownCount != expected.UnknownCount {
		t.Errorf("got UnknownCount %v expected %v", actual.UnknownCount, expected.UnknownCount)
	}

	for i := 0; i < len(actual.Slots); i++ {
		compareSlot(t, &actual.Slots[i], &expected.Slots[i], i)
	}
}

func compareSlot(t *testing.T, slot, testSlot *UserAreaSlot, i int) {
	if slot.Unknown != testSlot.Unknown {
		t.Errorf("got Slot[%d].Unknown %v expected %v", i, slot.Unknown, testSlot.Unknown)
	}

	if slot.ItemsCount != testSlot.ItemsCount {
		t.Errorf("got Slot[%d].ItemsCount %v expected %v", i, slot.ItemsCount, testSlot.ItemsCount)
	}

	if slot.User != testSlot.User {
		t.Errorf("got Slot[%d].User %v expected %v", i, slot.User, testSlot.User)
	}

	for j := 0; j < len(slot.Items); j++ {
		if slot.Items[j] != testSlot.Items[j] {
			t.Errorf("got Slot[%d].Items[%d] %v expected %v", i, j, slot.Items[j], testSlot.Items[j])
		}
	}
}
