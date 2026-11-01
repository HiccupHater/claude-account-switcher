package main

// Fork: the box that asks for an account's new name.

import (
	"log"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procGetDlgItemText = user32.NewProc("GetDlgItemTextW")
	procSendDlgItemMsg = user32.NewProc("SendDlgItemMessageW")
	procGetDlgItem     = user32.NewProc("GetDlgItem")
	procSetFocus       = user32.NewProc("SetFocus")
	nameDialogProc     = windows.NewCallback(nameDialog)
	// The one name box that can be open at a time and what it shows.
	nameMu               sync.Mutex
	nameInitial, nameOut string
)

const nameEdit = 100

// askName shows a box with prompt and an edit field holding current, and
// returns what was typed and whether OK was pressed. A box that is already
// open makes a second one return straight away.
func askName(title, prompt, current string) (string, bool) {
	if !nameMu.TryLock() {
		return "", false
	}
	defer nameMu.Unlock()
	done := make(chan uintptr)
	go func() {
		// The dialog's window belongs to this thread.
		runtime.LockOSThread()
		done <- showNameDialog(title, prompt, current)
	}()
	if <-done != idOK {
		return "", false
	}
	return nameOut, true
}

func showNameDialog(title, prompt, current string) uintptr {
	nameInitial, nameOut = current, ""
	const (
		wsPopup, wsCaption, wsSysMenu = 0x80000000, 0x00C00000, 0x00080000
		wsChild, wsVisible, wsTabStop = 0x40000000, 0x10000000, 0x00010000
		wsBorder                      = 0x00800000
		dsSetFont, dsModalFrame       = 0x40, 0x80
		dsSetForeground, dsCenter     = 0x200, 0x800
		wsExTopmost                   = 0x8
		bsDefPushButton               = 0x1
		esAutoHScroll                 = 0x80
		ssNoPrefix                    = 0x80
		classButton, classEdit        = 0x80, 0x81
		classStatic                   = 0x82
		width, margin, btnW, btnH     = 230, 7, 50, 14
		textHeight, editHeight        = 18, 12
	)
	editY := margin + textHeight + 2
	buttonsY := editY + editHeight + 8
	height := buttonsY + btnH + margin
	var t dialogTemplate
	t.header(wsPopup|wsCaption|wsSysMenu|dsSetFont|dsModalFrame|dsSetForeground|dsCenter, wsExTopmost,
		4, width, height, title, 9, "Segoe UI")
	t.item(wsChild|wsVisible|ssNoPrefix, margin, margin, width-2*margin, textHeight, 0xFFFF, classStatic, prompt)
	t.item(wsChild|wsVisible|wsTabStop|wsBorder|esAutoHScroll, margin, editY, width-2*margin, editHeight, nameEdit, classEdit, "")
	t.item(wsChild|wsVisible|wsTabStop|bsDefPushButton, width-margin-2*btnW-4, buttonsY, btnW, btnH, idOK, classButton, "Rename")
	t.item(wsChild|wsVisible|wsTabStop, width-margin-btnW, buttonsY, btnW, btnH, idCancel, classButton, "Cancel")

	var inst windows.Handle
	_ = windows.GetModuleHandleEx(0, nil, &inst)
	tmpl := t.aligned()
	r, _, err := procDialogBoxIndirectParam.Call(uintptr(inst), uintptr(unsafe.Pointer(&tmpl[0])), 0, nameDialogProc, 0)
	runtime.KeepAlive(tmpl)
	if int(r) <= 0 {
		log.Printf("name dialog: %v", err)
		return idCancel
	}
	return r
}

func nameDialog(hwnd, msg, wparam, _ uintptr) uintptr {
	const emSetSel, emLimitText = 0xB1, 0xC5
	switch msg {
	case wmInitDialog:
		setDialogText(hwnd, nameEdit, nameInitial)
		procSendDlgItemMsg.Call(hwnd, nameEdit, emLimitText, maxNickname, 0)
		procSendDlgItemMsg.Call(hwnd, nameEdit, emSetSel, 0, ^uintptr(0))
		edit, _, _ := procGetDlgItem.Call(hwnd, nameEdit)
		procSetFocus.Call(edit)
		// Focus was set by hand, so Windows must not move it.
		return 0
	case wmCommand:
		switch id := wparam & 0xFFFF; id {
		case idOK:
			buf := make([]uint16, maxNickname*2+1)
			procGetDlgItemText.Call(hwnd, nameEdit, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
			nameOut = windows.UTF16ToString(buf)
			procEndDialog.Call(hwnd, idOK)
		case idCancel:
			procEndDialog.Call(hwnd, idCancel)
		}
		return 1
	}
	return 0
}
