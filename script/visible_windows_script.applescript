tell application "System Events"
	set listOfProcesses to (every process whose visible is true)
end tell
repeat with proc in listOfProcesses
	set procName to (name of proc)
	set procID to (id of proc)
	log "PROCESS " & procID & ":" & procName
	set app_windows to (every window of proc)
	repeat with each_window in app_windows
		log "WINDOW -1:" & (name of each_window) as string
	end repeat
end repeat
