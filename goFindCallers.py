from collections import defaultdict
import os
import re

import sublime
import sublime_plugin
import os
import sys
import subprocess
import linecache
from threading import Thread

class GoFindCallersCommand(sublime_plugin.TextCommand):
	def run(self, edit):
		selections = self.view.sel()
		# Startup info stuff
		startupinfo = subprocess.STARTUPINFO()
		startupinfo.dwFlags |= subprocess.STARTF_USESHOWWINDOW

		plugPath = sublime.packages_path()
		# Open subprocess
		self.p = subprocess.Popen([plugPath+"\GoFindCallers\goFindCallers.exe"], startupinfo=startupinfo, stdout=subprocess.PIPE, stderr=subprocess.PIPE, stdin=subprocess.PIPE)

		self._callbackWithWordToFind(self._doFind)

	def _callbackWithWordToFind(self, callback):
		if len(self.view.sel()) == 1 and self.view.sel()[0].size() > 0:
			callback(self.view.substr(self.view.sel()[0]))
		else:
			sublime.status_message('Warnaing! No selection made')
			# self.view.window().show_input_panel('Enter word to find:', '', callback, None, None)


	def _doFind(self, wordToFind):
		toFind = re.escape(wordToFind)
		# print self.view.rowcol(self.view.sel()[0].begin())

		parsedLocations, err = self.p.communicate(wordToFind+"="+ self.view.file_name())
		if err != None:
			print err
			
		parsedLocations = parsedLocations.rstrip('\n')
		if parsedLocations == 'NotFound':
			sublime.status_message('Warnaing! Selection not found')
			return False

		# Create Find Results window
		self.resultsPane = self._getResultsPane()
		if self.resultsPane is None:
			return False

		parsedLines = parsedLocations.split('\n')
		toAppend = ["Matched " + str(len(parsedLines)/2) + " files for " + wordToFind]
		i = 0
		while i<(len(parsedLines)-1):
			fileLoc = parsedLines[i]
			regions = parsedLines[i+1].split(',')
			i = i + 2 
			filelines = linecache.getlines(fileLoc)
			linecount = len(filelines)

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


			toAppend.append("\n\n" + fileLoc + ":")
			# # append Lines
			lastLine = None
			for lineNumber in lines:
				if lastLine and lineNumber > lastLine + 1:
					toAppend.append(' ' * (5 - len(str(lineNumber)))
							+ '.' * len(str(lineNumber)))
				lastLine = lineNumber

				lineText = filelines[lineNumber-1].rstrip('\n')
				toAppend.append(self._format('%s' % (lineNumber),lineText, (str(lineNumber) in regions)))

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
		spacer = ' ' * (4 - len(str(lineNumber)))
		colon = ':' if match else ' '
		
		return ' {sp}{lineNumber}{colon} {text}'.format(lineNumber = lineNumber,
				colon = colon, text = line, sp = spacer)


	def _lineCount(self):
		return self.view.rowcol(self.view.size())[0]


class ShowResultsCommand(sublime_plugin.TextCommand):
	def run(self, edit, toAppend, toHighlight):
		self.view.erase(edit, sublime.Region(0, self.view.size()))
		self.view.insert(edit, self.view.size(), '\n'.join(toAppend))

		regions = self.view.find_all(toHighlight)
		self.view.add_regions('find_results', regions, 'found', '', 
				sublime.DRAW_OUTLINED)
