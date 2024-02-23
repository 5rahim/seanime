package util

import "testing"

func TestHandlePanicInModuleThen(t *testing.T) {

	type testStruct struct {
		mediaId int
	}

	testDangerousWork := func(obj *testStruct, work func()) {
		defer HandlePanicInModuleThen("util/panic_test", func() {
			obj.mediaId = 0
		})

		work()
	}

	var testCases = []struct {
		name            string
		obj             testStruct
		work            func()
		expectedMediaId int
	}{
		{
			name: "Test 1",
			obj:  testStruct{mediaId: 1},
			work: func() {
				panic("Test 1")
			},
			expectedMediaId: 0,
		},
		{
			name: "Test 2",
			obj:  testStruct{mediaId: 2},
			work: func() {
				// Do nothing
			},
			expectedMediaId: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			testDangerousWork(&tc.obj, tc.work)

			if tc.obj.mediaId != tc.expectedMediaId {
				t.Errorf("Expected mediaId to be %d, got %d", tc.expectedMediaId, tc.obj.mediaId)
			}

		})
	}

}
