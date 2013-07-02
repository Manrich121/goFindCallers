import re
import os

import sublime
import sublime_plugin
import subprocess
import linecache
from sys import platform as _platform

class GoFindCallersCommand(sublime_plugin.TextCommand):
	def run(self, edit):
		selections = self.view.sel()
		self.edit = edit	
		startupinfo = None
		plugPath = sublime.packages_path()
		# Check OS and build the executable path
		if _platform == "win32":
			# Startup info stuff, to block cmd window flash
			startupinfo = subprocess.STARTUPINFO()
			startupinfo.dwFlags |= subprocess.STARTF_USESHOWWINDOW
			# Windows
			processPath = plugPath+r'\GoFindCallers\bin\goFindCallers.exe'
		else:
			# linux and OS X
			processPath = plugPath+r'\GoFindCallers\bin\goFindCallers'  
			
		# Check exe build
		if not os.path.isfile(processPath):
			buildpath = plugPath+r'\GoFindCallers\src\cmd\goFindCallers'
			sublime.status_message('Installing plugin dependencies...')
			subprocess.Popen(["go", "install"], cwd=buildpath, startupinfo=startupinfo).wait()

		# Get gopath and format for stdin
		self.gopath = self.getenv()
		# Open subprocess
		self.p = subprocess.Popen([processPath], startupinfo=startupinfo, stdout=subprocess.PIPE, stderr=subprocess.PIPE, stdin=subprocess.PIPE)

		self._callbackWithWordToFind(self._doFind)

	def _callbackWithWordToFind(self, callback):
		if len(self.view.sel()) == 1 and self.view.sel()[0].size() > 0:
			callback(self.view.substr(self.view.sel()[0]))
		else:
			# sublime.status_message('Warnaing! No selection made')
			self.view.window().show_input_panel('Enter function to find:', '', callback, None, None)


	def _doFind(self, wordToFind):
		toFind = re.escape(wordToFind)
		self.count = 0
		uout = (wordToFind+"="+ self.view.file_name()+os.pathsep+self.gopath).encode("utf-8")
		print uout
		parsedLocations, err = self.p.communicate(uout)
		if err != None:
			print err
			
		parsedLocations = parsedLocations.rstrip('\n')
		if parsedLocations == 'NotFound':
			sublime.status_message('Warnaing! Selection not found')

		# Create Find Results window
		self.resultsPane = self._getResultsPane()
		if self.resultsPane is None:
			return False

		parsedLines = parsedLocations.split('\n')
		self.numFiles = len(parsedLines)/2
		toAppend = ["Matched " + unicode(self.numFiles) + " files for " + wordToFind]
		i = 0
		while i<(len(parsedLines)-1):
			fileLoc = parsedLines[i]
			regions = parsedLines[i+1].split(',')
			self.count += len(regions)
			i = i + 2 
			filelines = linecache.getlines(fileLoc)
			linecount = len(filelines)

			linecache.clearcache()
			# Determine the range of the lines across all regions
			lines = []
			prevNum = 0
			for r in regions:
				line = int(r)
				maxNum = min(linecount+1, line + 3)
				for curNum in range(max(0, line - 2), maxNum):
					if not curNum in lines:
						lines.append(curNum)
				prevNum = maxNum

			try:
				self.resultsPane.insert(self.edit, self.resultsPane.size(), '\n'.join(toAppend))
			except:
				print 'insert + decode'
				
			toAppend = ["\n\n" + fileLoc + ":"]
			# # append Lines
			lastLine = None
			for lineNumber in lines:
				if lastLine and lineNumber > lastLine + 1:
					toAppend.append(' ' * (5 - len(unicode(lineNumber)))
							+ '.' * len(unicode(lineNumber)))
				lastLine = lineNumber

				lineText = filelines[lineNumber-1].rstrip('\n').decode("utf-8")
				toAppend.append(self._format('%s' % (lineNumber),lineText, (unicode(lineNumber) in regions)))

		toAppend.append("\n"+ unicode(self.count) + " matches across " + unicode(self.numFiles) + " file(s)")
		self.resultsPane.run_command('show_results', {'toAppend': toAppend, 'toHighlight': toFind})

		self.view.window().focus_view(self.resultsPane)
		if self.p.poll() is None:
			self.p.terminate()

	# Borrowed from QuickSearch https://github.com/justinfoote/QuickSearch[.git]
	def _getResultsPane(self):
		"""Returns the results pane; creating one if necessary
		"""
		window = self.view.window()

		resultsPane = [v for v in window.views() 
			if v.name() == 'Find Results']

		if resultsPane:
			v = resultsPane[0]
		else:
			v = window.new_file()
			v.set_name('Find Results')
			v.settings().set('syntax', os.path.join(
					sublime.packages_path(), 'Default', 
					'Find Results.hidden-tmLanguage'))
			v.settings().set('draw_indent_guides', False)
			v.set_scratch(True)

		# window.focus_view(v)
		window.focus_view(self.view)
		return v


	def _format(self, lineNumber, line, match = False):
		spacer = ' ' * (4 - len(unicode(lineNumber)))
		colon = ':' if match else ' '
		
		return u' {sp}{lineNumber}{colon} {text}'.format(lineNumber = lineNumber,
				colon = colon, text = line, sp = spacer)

	def getenv(self):
		e = os.environ.copy()
		roots = e.get('GOPATH', '')
		return roots


class ShowResultsCommand(sublime_plugin.TextCommand):
	def run(self, edit, toAppend, toHighlight):
		# self.view.erase(edit, sublime.Region(0, self.view.size()))
		self.view.insert(edit, self.view.size(), '\n'.join(toAppend))

		regions = self.view.find_all(toHighlight)
		self.view.add_regions('find_results', regions, 'found', '', sublime.DRAW_OUTLINED)


# FindInFiles edited to handle double clicking and cntr+enter shortcuts
class FindInFilesGotoCommand(sublime_plugin.TextCommand):
    def run_(self, args):
        view = self.view
        if view.name() == "Find Results":
            line_no = self.get_line_no()
            file_name = self.get_file()
            if line_no is not None and file_name is not None:
                file_loc = "%s:%s" % (file_name, line_no)
                view.window().open_file(file_loc, sublime.ENCODED_POSITION)
            elif file_name is not None:
                view.window().open_file(file_name)
        else:
            system_command = args["command"] if "command" in args else None
            if system_command:
                system_args = dict({"event": args["event"]}.items() + args["args"].items())
                self.view.run_command(system_command, system_args)

    def get_line_no(self):
        view = self.view
        if len(view.sel()) == 1:
            line_text = view.substr(view.line(view.sel()[0]))
            match = re.match(r"\s*(\d+).+", line_text)
            if match:
                return match.group(1)
        return None

    def get_file(self):
        view = self.view
        if len(view.sel()) == 1:
            line = view.line(view.sel()[0])
            while line.begin() > 0:
                line_text = view.substr(line)
                match = re.match(r"(.+):$", line_text)
                if match:
                    if os.path.exists(match.group(1)):
                        return match.group(1)
                line = view.line(line.begin() - 1)
        return None
