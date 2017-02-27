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
	area1.Revision = 2
	slot1 := &area1.Slots[1]
	slot1.IsFree = 0
	for i := 0; i < len(slot1.Items); i++ {
		slot1.Items[i] = NoUserAreaItem
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
		t.Errorf("got IsFree %d expected %d", area.Unknown, expectedUnknown)
	}

	const expectedUnknownCount = 7
	if area.Revision != expectedUnknownCount {
		t.Errorf("got Revision %d expected %d", area.Revision, expectedUnknownCount)
	}

	slot0 := area.Slots[0]
	expectedSlot0 := UserAreaSlot{
		IsFree: 1,
		Items: [10]UserAreaItem{
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
		},
		ItemsCount: 0,
		User:       0,
	}

	compareSlot(t, &slot0, &expectedSlot0, 0)

	slot5 := area.Slots[5]
	expectedSlot5 := UserAreaSlot{
		IsFree: 0,
		Items: [10]UserAreaItem{
			UserAreaItem(26),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
			UserAreaItem(NoUserAreaItem),
		},
		ItemsCount: 1,
		User:       76561198028565520,
	}

	compareSlot(t, &slot5, &expectedSlot5, 05)

	return
}

func compareArea(t *testing.T, actual, expected *UserArea) {
	if actual.Unknown != expected.Unknown {
		t.Errorf("got IsFree %v expected %v", actual.Unknown, expected.Unknown)
	}

	if actual.Revision != expected.Revision {
		t.Errorf("got Revision %v expected %v", actual.Revision, expected.Revision)
	}

	for i := 0; i < len(actual.Slots); i++ {
		compareSlot(t, &actual.Slots[i], &expected.Slots[i], i)
	}
}

func compareSlot(t *testing.T, slot, testSlot *UserAreaSlot, i int) {
	if slot.IsFree != testSlot.IsFree {
		t.Errorf("got Slot[%d].IsFree %v expected %v", i, slot.IsFree, testSlot.IsFree)
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
