package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	varibale "github.com/yorukot/superfile/src/config"
)

// Change file panel mode (select mode or browser mode)
func changeFilePanelMode(m model) model {
	panel := m.fileModel.filePanels[m.filePanelFocusIndex]
	if panel.panelMode == selectMode {
		panel.selected = panel.selected[:0]
		panel.panelMode = browserMode
	} else if panel.panelMode == browserMode {
		panel.panelMode = selectMode
	}
	m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	return m
}

// Back to parent directory
func parentDirectory(m model) model {
	panel := m.fileModel.filePanels[m.filePanelFocusIndex]
	panel.directoryRecord[panel.location] = directoryRecord{
		directoryCursor: panel.cursor,
		directoryRender: panel.render,
	}
	fullPath := panel.location
	parentDir := path.Dir(fullPath)
	panel.location = parentDir
	directoryRecord, hasRecord := panel.directoryRecord[panel.location]
	if hasRecord {
		panel.cursor = directoryRecord.directoryCursor
		panel.render = directoryRecord.directoryRender
	} else {
		panel.cursor = 0
		panel.render = 0
	}
	m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	return m
}

// Enter director or open file with default application
func enterPanel(m model) model {
	panel := m.fileModel.filePanels[m.filePanelFocusIndex]

	if len(panel.element) == 0 {
		return m
	}

	if panel.element[panel.cursor].directory {
		panel.directoryRecord[panel.location] = directoryRecord{
			directoryCursor: panel.cursor,
			directoryRender: panel.render,
		}
		panel.location = panel.element[panel.cursor].location
		directoryRecord, hasRecord := panel.directoryRecord[panel.location]
		if hasRecord {
			panel.cursor = directoryRecord.directoryCursor
			panel.render = directoryRecord.directoryRender
		} else {
			panel.cursor = 0
			panel.render = 0
		}
		panel.searchBar.SetValue("")
	} else if !panel.element[panel.cursor].directory {
		fileInfo, err := os.Lstat(panel.element[panel.cursor].location)
		if err != nil {
			outPutLog("err when getting file info", err)
			return m
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			if isBrokenSymlink(panel.element[panel.cursor].location) {
				return m
			}

			linkPath, _ := os.Readlink(panel.element[panel.cursor].location)

			absLinkPath, err := filepath.Abs(linkPath)
			if err != nil {
				return m
			}

			m.fileModel.filePanels[m.filePanelFocusIndex].location = absLinkPath
			return m
		}

		openCommand := "xdg-open"
		if runtime.GOOS == "darwin" {
			openCommand = "open"
		}
		cmd := exec.Command(openCommand, panel.element[panel.cursor].location)
		_, err = cmd.Output()
		if err != nil {
			outPutLog(fmt.Sprintf("err when open file with %s", openCommand), err)
		}

	}

	m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	return m
}

// Switch to the directory where the sidebar cursor is located
func sidebarSelectDirectory(m model) model {
	m.focusPanel = nonePanelFocus
	panel := m.fileModel.filePanels[m.filePanelFocusIndex]

	panel.directoryRecord[panel.location] = directoryRecord{
		directoryCursor: panel.cursor,
		directoryRender: panel.render,
	}

	panel.location = m.sidebarModel.directories[m.sidebarModel.cursor].location
	directoryRecord, hasRecord := panel.directoryRecord[panel.location]
	if hasRecord {
		panel.cursor = directoryRecord.directoryCursor
		panel.render = directoryRecord.directoryRender
	} else {
		panel.cursor = 0
		panel.render = 0
	}
	panel.focusType = focus
	m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	return m
}

// Select all item in the file panel (only work on select mode)
func selectAllItem(m model) model {
	panel := m.fileModel.filePanels[m.filePanelFocusIndex]
	for _, item := range panel.element {
		panel.selected = append(panel.selected, item.location)
	}
	m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	return m
}

// Select the item where cursor located (only work on select mode)
func singleItemSelect(m model) model {
	panel := m.fileModel.filePanels[m.filePanelFocusIndex]
	if arrayContains(panel.selected, panel.element[panel.cursor].location) {
		panel.selected = removeElementByValue(panel.selected, panel.element[panel.cursor].location)
	} else {
		panel.selected = append(panel.selected, panel.element[panel.cursor].location)
	}

	m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	return m
}

// Toggle dotfile display or not
func toggleDotFileController(m model) model {
	newToggleDotFile := ""
	if m.toggleDotFile {
		newToggleDotFile = "false"
		m.toggleDotFile = false
	} else {
		newToggleDotFile = "true"
		m.toggleDotFile = true
	}
	err := os.WriteFile(varibale.ToggleDotFilea, []byte(newToggleDotFile), 0644)
	if err != nil {
		outPutLog("Pinned folder function updatedData superfile data error", err)
	}

	return m
}

// Focus on search bar
func searchBarFocus(m model) model {
	panel := m.fileModel.filePanels[m.filePanelFocusIndex]
	if panel.searchBar.Focused() {
		panel.searchBar.Blur()
	} else {
		panel.searchBar.Focus()
	}

	// config search bar width
	panel.searchBar.Width = m.fileModel.width - 4
	m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	return m
}

// ======================================== File panel controller ========================================

// Control file panel list up
func controlFilePanelListUp(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.fileModel.filePanels[m.filePanelFocusIndex]
		if len(panel.element) == 0 {
			return m
		}
		if panel.cursor > 0 {
			panel.cursor--
			if panel.cursor < panel.render {
				panel.render--
			}
		} else {
			if len(panel.element) > panelElementHeight(m.mainPanelHeight) {
				panel.render = len(panel.element) - panelElementHeight(m.mainPanelHeight)
				panel.cursor = len(panel.element) - 1
			} else {
				panel.cursor = len(panel.element) - 1
			}
		}

		m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	}
	return m
}

// Control file panel list down
func controlFilePanelListDown(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.fileModel.filePanels[m.filePanelFocusIndex]
		if len(panel.element) == 0 {
			return m
		}
		if panel.cursor < len(panel.element)-1 {
			panel.cursor++
			if panel.cursor > panel.render+panelElementHeight(m.mainPanelHeight)-1 {
				panel.render++
			}
		} else {
			panel.render = 0
			panel.cursor = 0
		}
		m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	}

	return m
}

// Handles the action of selecting an item in the file panel upwards. (only work on select mode)
func itemSelectUp(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.fileModel.filePanels[m.filePanelFocusIndex]
		if panel.cursor > 0 {
			panel.cursor--
			if panel.cursor < panel.render {
				panel.render--
			}
		} else {
			if len(panel.element) > panelElementHeight(m.mainPanelHeight) {
				panel.render = len(panel.element) - panelElementHeight(m.mainPanelHeight)
				panel.cursor = len(panel.element) - 1
			} else {
				panel.cursor = len(panel.element) - 1
			}
		}
		if arrayContains(panel.selected, panel.element[panel.cursor].location) {
			panel.selected = removeElementByValue(panel.selected, panel.element[panel.cursor].location)
		} else {
			panel.selected = append(panel.selected, panel.element[panel.cursor].location)
		}

		m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	}
	return m
}

// Handles the action of selecting an item in the file panel downwards. (only work on select mode)
func itemSelectDown(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.fileModel.filePanels[m.filePanelFocusIndex]
		if panel.cursor < len(panel.element)-1 {
			panel.cursor++
			if panel.cursor > panel.render+panelElementHeight(m.mainPanelHeight)-1 {
				panel.render++
			}
		} else {
			panel.render = 0
			panel.cursor = 0
		}
		if arrayContains(panel.selected, panel.element[panel.cursor].location) {
			panel.selected = removeElementByValue(panel.selected, panel.element[panel.cursor].location)
		} else {
			panel.selected = append(panel.selected, panel.element[panel.cursor].location)
		}

		m.fileModel.filePanels[m.filePanelFocusIndex] = panel
	}
	return m
}

// ======================================== Sidebar controller ========================================

// Yorukot: P.S God bless me, this sidebar controller code is really ugly...

// Control sidebar panel list up
func controlSideBarListUp(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.sidebarModel.cursor > 0 {
			m.sidebarModel.cursor--
		} else {
			m.sidebarModel.cursor = len(m.sidebarModel.directories) - 1
		}
		newDirectory := m.sidebarModel.directories[m.sidebarModel.cursor].location

		for newDirectory == "Pinned+-*/=?" || newDirectory == "Disks+-*/=?" {
			m.sidebarModel.cursor--
			newDirectory = m.sidebarModel.directories[m.sidebarModel.cursor].location
		}
		changeToPlus := false
		cursorRender := false
		for !cursorRender {
			totalHeight := 2
			for i := m.sidebarModel.renderIndex; i < len(m.sidebarModel.directories); i++ {
				if totalHeight >= m.mainPanelHeight {
					break
				}
				directory := m.sidebarModel.directories[i]

				if directory.location == "Pinned+-*/=?" {
					totalHeight += 3
					continue
				}

				if directory.location == "Disks+-*/=?" {
					if m.mainPanelHeight-totalHeight <= 2 {
						break
					}
					totalHeight += 3
					continue
				}

				totalHeight++
				if m.sidebarModel.cursor == i && m.focusPanel == sidebarFocus {
					cursorRender = true
				}
			}

			if changeToPlus {
				m.sidebarModel.renderIndex++
				continue
			}

			if !cursorRender {
				m.sidebarModel.renderIndex--
			}
			if m.sidebarModel.renderIndex < 0 {
				changeToPlus = true
				m.sidebarModel.renderIndex++
			}
		}

		if changeToPlus {
			m.sidebarModel.renderIndex--
		}
	}
	return m
}

// Control sidebar panel list down
func controlSideBarListDown(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		lenDirs := len(m.sidebarModel.directories)
		if m.sidebarModel.cursor < lenDirs-1 {
			m.sidebarModel.cursor++
		} else {
			m.sidebarModel.cursor = 0
		}

		newDirectory := m.sidebarModel.directories[m.sidebarModel.cursor].location
		for newDirectory == "Pinned+-*/=?" || newDirectory == "Disks+-*/=?" {
			m.sidebarModel.cursor++
			if m.sidebarModel.cursor+1 >= len(m.sidebarModel.directories) {
				m.sidebarModel.cursor = 0
			}
			newDirectory = m.sidebarModel.directories[m.sidebarModel.cursor].location
		}
		cursorRender := false
		for !cursorRender {
			totalHeight := 2
			for i := m.sidebarModel.renderIndex; i < len(m.sidebarModel.directories); i++ {
				if totalHeight >= m.mainPanelHeight {
					break
				}

				directory := m.sidebarModel.directories[i]

				if directory.location == "Pinned+-*/=?" {
					totalHeight += 3
					continue
				}

				if directory.location == "Disks+-*/=?" {
					if m.mainPanelHeight-totalHeight <= 2 {
						break
					}
					totalHeight += 3
					continue
				}

				totalHeight++
				if m.sidebarModel.cursor == i && m.focusPanel == sidebarFocus {
					cursorRender = true
				}
			}

			if !cursorRender {
				m.sidebarModel.renderIndex++
			}
			if m.sidebarModel.renderIndex > m.sidebarModel.cursor {
				m.sidebarModel.renderIndex = 0
			}
		}
	}
	return m
}

// ======================================== Metadata controller ========================================

// Control metadata panel up
func controlMetadataListUp(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.fileMetaData.renderIndex > 0 {
			m.fileMetaData.renderIndex--
		} else {
			m.fileMetaData.renderIndex = len(m.fileMetaData.metaData) - 1
		}
	}
	return m
}

// Control metadata panel down
func controlMetadataListDown(m model, wheel bool) model {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.fileMetaData.renderIndex < len(m.fileMetaData.metaData)-1 {
			m.fileMetaData.renderIndex++
		} else {
			m.fileMetaData.renderIndex = 0
		}
	}
	return m
}

// ======================================== Processbar controller ========================================

// Control processbar panel list up
func controlProcessbarListUp(m model, wheel bool) model {
	if len(m.processBarModel.processList) == 0 {
		return m
	}
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.processBarModel.cursor > 0 {
			m.processBarModel.cursor--
			if m.processBarModel.cursor < m.processBarModel.render {
				m.processBarModel.render--
			}
		} else {
			if len(m.processBarModel.processList) <= 3 || (len(m.processBarModel.processList) <= 2 && footerHeight < 14) {
				m.processBarModel.cursor = len(m.processBarModel.processList) - 1
			} else {
				m.processBarModel.render = len(m.processBarModel.processList) - 3
				m.processBarModel.cursor = len(m.processBarModel.processList) - 1
			}
		}
	}
	return m
}

// Control processbar panel list down
func controlProcessbarListDown(m model, wheel bool) model {
	if len(m.processBarModel.processList) == 0 {
		return m
	}

	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.processBarModel.cursor < len(m.processBarModel.processList)-1 {
			m.processBarModel.cursor++
			if m.processBarModel.cursor > m.processBarModel.render+2 {
				m.processBarModel.render++
			}
		} else {
			m.processBarModel.render = 0
			m.processBarModel.cursor = 0
		}
	}
	return m
}
