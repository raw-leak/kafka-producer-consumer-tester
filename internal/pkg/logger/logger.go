package logger

import (
	"fmt"
	"strings"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type MessageStats struct {
	InProgress int
	Success    int
	Failed     int
}

type Logger struct {
	mutex  sync.Mutex
	ticker *time.Ticker

	msgsTable *widgets.Table
	sent      MessageStats
	processed MessageStats

	partTable  *widgets.Table
	partitions int
	processors int

	logsList *widgets.List
}

func New() (*Logger, error) {
	if err := ui.Init(); err != nil {
		return nil, err
	}

	grid := ui.NewGrid()
	termWidth, _ := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, 30)

	// Message verification table
	msgsTable := widgets.NewTable()
	msgsTable.Title = "Message Processing Status"
	msgsTable.Rows = [][]string{
		{"Status", "In-progress", "Success", "Failed"},
		{"Sent Events", "0", "0", "0"},
		{"Processed Events", "0", "0", "0"},
	}
	msgsTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	msgsTable.TextAlignment = ui.AlignCenter
	msgsTable.RowSeparator = true
	msgsTable.BorderStyle = ui.NewStyle(ui.ColorCyan)

	// Partitions table
	partTable := widgets.NewTable()
	partTable.Title = "Partitions/Processors Status"
	partTable.Rows = [][]string{
		{"Entity", "Working"},
		{"Partitions", "0"},
		{"Processors", "0"},
	}
	partTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	partTable.TextAlignment = ui.AlignCenter
	partTable.RowSeparator = true
	partTable.BorderStyle = ui.NewStyle(ui.ColorCyan)

	// General logs table
	logsList := widgets.NewList()
	logsList.Title = "Logs"
	logsList.Rows = []string{}
	logsList.TextStyle = ui.NewStyle(ui.ColorYellow)
	logsList.WrapText = false

	// Render the defined elements

	// Calculate the required height for the message table
	msgsHeight := len(msgsTable.Rows) + 1
	partHeight := len(partTable.Rows) + 1
	logsHeight := len(logsList.Rows) + 1

	totalHeight := msgsHeight + partHeight + logsHeight

	grid.Set(
		ui.NewRow(2.3/float64(totalHeight), ui.NewCol(1.0, msgsTable)),
		ui.NewRow(2.3/float64(totalHeight), ui.NewCol(1.0, partTable)),
		ui.NewRow(10/float64(totalHeight), ui.NewCol(1.0, logsList)),
	)
	ui.Render(grid)

	// Refresh rate
	ticker := time.NewTicker(time.Millisecond * 500)

	logger := &Logger{msgsTable: msgsTable, ticker: ticker, logsList: logsList, partTable: partTable}

	go logger.run()

	return logger, nil
}

func (l *Logger) updateMsgsTable() {
	l.msgsTable.Rows[1][1] = fmt.Sprintf("%d", l.sent.InProgress)
	l.msgsTable.Rows[1][2] = fmt.Sprintf("%d", l.sent.Success)
	l.msgsTable.Rows[1][3] = fmt.Sprintf("%d", l.sent.Failed)
	l.msgsTable.Rows[2][1] = fmt.Sprintf("%d", l.processed.InProgress)
	l.msgsTable.Rows[2][2] = fmt.Sprintf("%d", l.processed.Success)
	l.msgsTable.Rows[2][3] = fmt.Sprintf("%d", l.processed.Failed)
	ui.Render(l.msgsTable)
}

func (l *Logger) updateLogList() {
	ui.Render(l.logsList)
}

func (l *Logger) updatePartTable() {
	l.partTable.Rows[1][1] = fmt.Sprintf("%d", l.partitions)
	l.partTable.Rows[2][1] = fmt.Sprintf("%d", l.processors)
	ui.Render(l.partTable)
}

func (l *Logger) run() {
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			if e.ID == "q" || e.ID == "<C-c>" {
				l.Shutdown()
				return
			}
		case <-l.ticker.C:
			l.updateMsgsTable()
			l.updateLogList()
			l.updatePartTable()
		}
	}
}

func (l *Logger) RecordSent(status string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	switch status {
	case "in-progress":
		l.sent.InProgress++
	case "success":
		l.sent.Success++
	case "failed":
		l.sent.Failed++
	}
}

func (l *Logger) RecordProcessed(status string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	switch status {
	case "in-progress":
		l.processed.InProgress++
	case "success":
		l.processed.Success++
	case "failed":
		l.processed.Failed++
	}
}

func (l *Logger) AddedPartition() {
	l.partitions++
}

func (l *Logger) RemovedPartition() {
	l.partitions--
}

func (l *Logger) AddedProcessor() {
	l.processors++
}

func (l *Logger) RemovedProcessor() {
	l.processors--
}

func (l *Logger) Info(msg string) {
	l.logsList.Rows = append(l.logsList.Rows, fmt.Sprintf("INFO: %s", msg))
}

func (l *Logger) Infof(format string, v ...any) {
	l.logsList.Rows = append(l.logsList.Rows, fmt.Sprintf("INFO: %s", fmt.Sprintf(format, v...)))
}

func (l *Logger) Error(msg string) {
	l.logsList.Rows = append(l.logsList.Rows, fmt.Sprintf("ERROR: %s", msg))
}

func (l *Logger) Errorf(format string, v ...any) {
	l.logsList.Rows = append(l.logsList.Rows, fmt.Sprintf("ERROR: %s", fmt.Sprintf(format, v...)))
}

func captureUIState(msgsTable *widgets.Table, partTable *widgets.Table, logsList *widgets.List) string {
	var builder strings.Builder

	// Function to capture table data
	captureTable := func(t *widgets.Table) {
		builder.WriteString(t.Title + "\n")
		for _, row := range t.Rows {
			builder.WriteString(strings.Join(row, " | ") + "\n")
		}
		builder.WriteString("\n") // Add a newline for spacing between tables
	}

	// Capture each table
	captureTable(msgsTable)
	captureTable(partTable)

	// Capture list data
	builder.WriteString(logsList.Title + "\n")
	for _, row := range logsList.Rows {
		builder.WriteString(row + "\n")
	}

	return builder.String()
}

func (l *Logger) Shutdown() {
	l.Info("the app will be closed in 10 seconds...")
	l.Info("Thank you!")

	time.Sleep(time.Second * 10)
	l.ticker.Stop()

	ui.Close()
}
