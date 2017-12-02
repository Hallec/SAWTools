#!/usr/bin/env python3

import secrets
import requests
from bs4 import BeautifulSoup as bs

#Terminal colors
colors = {
    "blue" : '\033[94m',
    "green" : '\033[92m',
    "yellow" : '\033[93m',
    "red": '\033[91m',
    "end": '\033[0m',
}


def random_color():
	choice = secrets.choice(range(len(colors)-2))
	
	for i,color in enumerate(colors):
		if i == choice: return colors[color]


#Introduction Title
print(random_color(),"""
 █████╗ ██████╗ ██╗  ██╗    ██████╗  ██████╗ ██╗    ██╗███╗   ██╗██╗      ██████╗  █████╗ ██████╗ ███████╗██████╗ 
██╔══██╗██╔══██╗██║ ██╔╝    ██╔══██╗██╔═══██╗██║    ██║████╗  ██║██║     ██╔═══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗
███████║██████╔╝█████╔╝     ██║  ██║██║   ██║██║ █╗ ██║██╔██╗ ██║██║     ██║   ██║███████║██║  ██║█████╗  ██████╔╝
██╔══██║██╔═══╝ ██╔═██╗     ██║  ██║██║   ██║██║███╗██║██║╚██╗██║██║     ██║   ██║██╔══██║██║  ██║██╔══╝  ██╔══██╗
██║  ██║██║     ██║  ██╗    ██████╔╝╚██████╔╝╚███╔███╔╝██║ ╚████║███████╗╚██████╔╝██║  ██║██████╔╝███████╗██║  ██║
╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝    ╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═╝  ╚═══╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝
""",colors["end"])

apk = input("Introduce the app name you want to download: ")

#Checking up the different options
response = requests.get("https://www.apkmirror.com/?post_type=app_release&searchtype=apk&s={}".format(apk.lower()))
html_response = bs(response.text,"html.parser")
res = html_response.find_all("div",{"class":"addpadding"})

#Function for listing every download option
def list_options(title,opt_list):
	valid = False
	while not valid:	
		print(title)
		for index in opt_list:
			print("{}. {}".format(index,opt_list[index]["name"]))
			
		try:
			opt = int(input("Introduce the app version, you want to download: "))
			if opt < 1 or opt > len(opt_list):
				if len(opt_list) == 1: print("[ERROR] The option must be: 1")
				print("[ERROR] The option must be between: 1 and {}".format(len(opt_list)))
			else:
				return opt
		except:
			print("[ERROR] The option chosen must be numeric".format(len(opt_list)))

#The app is found in the website
if not res:
	list_widget = html_response.find_all("div",{"class":"listWidget"})
	rows = list_widget[0].find_all("div",{"class":"appRow"})
	print("App found. {} {}.\n\n".format(len(rows),"results" if len(rows) > 0 else "result"))
	apks_list = {}
	for index,row in enumerate(rows):
		a_link = row.find("a")
		apk_name = a_link.text
		apk_link = a_link["href"]
		apks_list[index+1] = {"name": apk_name, "link": apk_link}
	
	
	#Listing every option found
	opt = list_options(title="--- APPS LIST ---",opt_list=apks_list)
	
	for index in apks_list:
		print("{}[{}]. {}{}".format(colors["green"] if index == opt else "","*" if index == opt else " ",apks_list[index]["name"],colors["end"] if index == opt else ""))
	print("\n"*2)
	print("Searching available architectures...")
	res_download = requests.get("https://www.apkmirror.com/{}".format(apks_list[opt]["link"]))
	res_download = bs(res_download.text,"html.parser")
	list_widget = res_download.find_all("div",{"class":"listWidget"})[0]
	rows = list_widget.find_all("div",{"class":"table-row headerFont"})
	
	archs_list = {}
	
	for index,row in enumerate(rows[1:]):
		a_link = row.find("a")
		a_name = a_link.text.strip()
		apk_link = a_link["href"]
		arch = row.find_all("div",{"class":"table-cell rowheight addseparator expand pad dowrap"})[1].text
		archs_list[index+1] = {"name": a_name, "link": apk_link, "arch": arch}
	
	#Listing hardware architectures available
	opt = list_options(title="--- ARCHITECTURES LIST ---",opt_list=archs_list)
	for index in archs_list:
		print("{}[{}]. {}{}".format(colors["green"] if index == opt else "","*" if index == opt else " ",archs_list[index]["name"],colors["end"] if index == opt else ""))
	print("\n"*2)
	res_apk = requests.get("https://www.apkmirror.com/{}".format(archs_list[opt]["link"]))
	res_apk = bs(res_apk.text,"html.parser")
	download_link = "https://www.apkmirror.com/{}".format(res_apk.find("a",{"class":"btn btn-flat downloadButton"})["href"])
	
	apk_name = input("Choose a name to save the file. (By omission: {}): ".format(apk))
	if apk_name:
		apk = apk_name
	
	
	#Downloading the Android app to input directory, in order to analyzing its manifest file	
	app_apk = requests.get(download_link,stream=True)

	try:
		with open("/home/saw/Scripts/input/{}.apk".format(apk), 'wb') as f:
			for chunk in app_apk.iter_content(chunk_size=1024):
		 		if chunk:
		 			f.write(chunk)
		print("* apk: {}.apk downloaded successfully.".format(apk))
	except:
		print("[ERROR] While downloading app. Retry again.")

#App not found in the website
else:
	print("APK not found. Please, try another Android app...")



