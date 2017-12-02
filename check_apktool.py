#!/usr/bin/env python3

import os
import re
import requests
import datetime
import subprocess
from collections import OrderedDict
from bs4 import BeautifulSoup as bs
import time

VERSIONS = OrderedDict()

#Terminal Colors
TERM_COLORS = {"red": "\033[91m",
	       "green":"\033[92m",
		"end": "\033[0m",}

#Function for displating the script name
def introduction():
	print(TERM_COLORS["green"],"""
	 █████╗ ██████╗ ██╗  ██╗████████╗ ██████╗  ██████╗ ██╗          ██████╗██╗  ██╗███████╗ ██████╗██╗  ██╗███████╗██████╗ 
	██╔══██╗██╔══██╗██║ ██╔╝╚══██╔══╝██╔═══██╗██╔═══██╗██║         ██╔════╝██║  ██║██╔════╝██╔════╝██║ ██╔╝██╔════╝██╔══██╗
	███████║██████╔╝█████╔╝    ██║   ██║   ██║██║   ██║██║         ██║     ███████║█████╗  ██║     █████╔╝ █████╗  ██████╔╝
	██╔══██║██╔═══╝ ██╔═██╗    ██║   ██║   ██║██║   ██║██║         ██║     ██╔══██║██╔══╝  ██║     ██╔═██╗ ██╔══╝  ██╔══██╗
	██║  ██║██║     ██║  ██╗   ██║   ╚██████╔╝╚██████╔╝███████╗    ╚██████╗██║  ██║███████╗╚██████╗██║  ██╗███████╗██║  ██║
	╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝   ╚═╝    ╚═════╝  ╚═════╝ ╚══════╝     ╚═════╝╚═╝  ╚═╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
	""",TERM_COLORS["end"])

#Function for getting the current time
def get_time():
	return datetime.datetime.now().strftime("%d/%m/%Y - %H:%M:%S")

#Function for moving the file to the correct bin folder
def mv(current_path,future_path):
	#subprocess.Popen(["mv",current_path,future_path],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL,shell=True)
	os.rename(current_path,future_path)	

#Function for removing files in temp folder
def rm(files):
	files_path = map(lambda f: "/home/saw/Scripts/tmp/{}".format(f),files) 
	for f in files_path:
		subprocess.Popen(["rm","-rf",f]) 

#Function for setting correct permissions for files a list of files
def chmod(permission,files):
	files_path = list(map(lambda f: "/home/saw/Scripts/tmp/{}".format(f),files))
	for cfile in files_path: os.chmod(cfile,permission)	

#Function for updating configuration files
def update_cfg_files():
	with open("/home/saw/Scripts/conf/versions.cfg","w") as f:
		for key in VERSIONS:
			f.write("{}={}\n".format(key,VERSIONS[key]))

#Checking apktool version installed in the system
def check_current_version():
	introduction()
	with open("/home/saw/Scripts/conf/versions.cfg","r") as f: 
		lines =	f.read().splitlines()

	for line in lines:
		software,version = line.split("=")
		VERSIONS[software.strip()] = version.strip()

	apktool_web = requests.get("https://bitbucket.org/iBotPeaches/apktool/downloads/").text
	html = bs(apktool_web,"html.parser")
	uploaded_files = html.find_all("table",{"id": "uploaded-files"})[0].find_all("td",{"class":"name"})
	latest_version = re.search(r"^.*_(.*)\.jar$",uploaded_files[1].find("a")["href"]).group(1)
        
	if latest_version > VERSIONS["APKTOOL_VERSION"]:
		print("[{}] Downloading wrapper script...".format(get_time()))
		wrapper = requests.get("https://raw.githubusercontent.com/iBotPeaches/Apktool/master/scripts/linux/apktool",
				     stream=True)
		with open("/home/saw/Scripts/tmp/apktool","wb+") as f:
			for chunk in wrapper.iter_content(chunk_size=1024):
				if chunk: f.write(chunk)
 
		print("[{}] Downloading latest version of apktool [{}]...".format(get_time(),latest_version))
		apkjar = requests.get("https://bitbucket.org/iBotPeaches/apktool/downloads/apktool_{}.jar".format(latest_version),
				     stream=True)
		with open("/home/saw/Scripts/tmp/apktool.jar","wb+") as f: 
			for chunk in apkjar.iter_content(chunk_size=1024):
				if chunk: f.write(chunk)
		print("[{}] Download process complete...".format(get_time()))
		print("[{}] Fixing permissions for files...".format(get_time()))
		chmod(0o755,["apktool","apktool.jar"])
		print("[{}] Moving apktool to bin directory...".format(get_time()))
		try:

			for cfile in ["apktool","apktool.jar"]: mv("/home/saw/Scripts/tmp/{}".format(cfile),"/usr/local/bin/{}".format(cfile))
		except Exception as e:
			print(e)
			print(TERM_COLORS["red"],
			     "* [ERROR] You need root permission to move files.",
			      TERM_COLORS["end"])
			print("[{}] Cleaning temp directory...".format(get_time()))
			rm(["apktool","apktool.jar"])
			exit(-1)
		print("[{}] Updating configuration files...".format(get_time()))
		print(TERM_COLORS["green"],
		      "[{}] apktool is ready to use.".format(get_time()),
		      TERM_COLORS["end"])
		VERSIONS["APKTOOL_VERSION"] = latest_version
		update_cfg_files()
	else:
		print(TERM_COLORS["green"],
		     "* [{}] You have installed the current version of apktool. No updates available.".format(get_time()),
		      TERM_COLORS["end"])
		
if __name__ == "__main__": check_current_version() 
