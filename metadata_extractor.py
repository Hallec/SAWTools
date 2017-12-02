#!/usr/bin/env python3

import os
import re
import time
import json
import secrets
import requests
import datetime
import argparse
import googlemaps
import subprocess
import configparser
from PIL import Image
from selenium import webdriver
from selenium.webdriver.chrome.options import Options

class MetadataExtractor:
    #Terminal colors
    colors = {
        "blue" : '\033[94m',
        "green" : '\033[92m',
        "yellow" : '\033[93m',
        "red": '\033[91m',
        "end": '\033[0m',
    }

    #Constructor
    def __init__(self):
        self.metadata = {}
        self.gps_tags = {}
        self.config = configparser.ConfigParser()
        self.config.read("/home/saw/Scripts/conf/api_keys.cfg")
        self.config.sections()

        #Enable key/value, gps and latitude/longitude detection
        self.__create_key_value_regex()
        self.__detect_location_tags()
        self.__extract_lat_long_values()

    #Function for selecting a random color for terminal messages
    def __random_color(self):
        choice = secrets.choice(range(len(self.colors)-2))
    
        for i,color in enumerate(self.colors):
            if i == choice: return self.colors[color]

    
    #Function for getting current date time
    def __get__time(self):
        return datetime.datetime.now().strftime("%A - %d-%m-%Y at %H:%M:%S").capitalize()

    #Function for checking Google Maps API key
    def __checking_api_key(self):
        try:
            self.api_key = self.config["KEYS"]["GoogleMaps"].strip()
            if self.api_key == "":
                print("[ERROR] The Google Maps API Key is empty. Please, introduce your key in order to use this service, (conf/api_keys.cfg)")
        except:
            exit()

    #Function for introducing the script
    def introduction(self):
        print(self.__random_color(),"""
        ███╗   ███╗███████╗████████╗ █████╗ ██████╗  █████╗ ████████╗ █████╗     ███████╗██╗  ██╗████████╗██████╗  █████╗  ██████╗████████╗ ██████╗ ██████╗ 
        ████╗ ████║██╔════╝╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗    ██╔════╝╚██╗██╔╝╚══██╔══╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗
        ██╔████╔██║█████╗     ██║   ███████║██║  ██║███████║   ██║   ███████║    █████╗   ╚███╔╝    ██║   ██████╔╝███████║██║        ██║   ██║   ██║██████╔╝
        ██║╚██╔╝██║██╔══╝     ██║   ██╔══██║██║  ██║██╔══██║   ██║   ██╔══██║    ██╔══╝   ██╔██╗    ██║   ██╔══██╗██╔══██║██║        ██║   ██║   ██║██╔══██╗
        ██║ ╚═╝ ██║███████╗   ██║   ██║  ██║██████╔╝██║  ██║   ██║   ██║  ██║    ███████╗██╔╝ ██╗   ██║   ██║  ██║██║  ██║╚██████╗   ██║   ╚██████╔╝██║  ██║
        ╚═╝     ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝    ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
            """,self.colors["end"])

    def check_input_dir(self):
        self.files = os.listdir("/home/saw/Scripts/input/metadata")
        if not self.files:
            print("[ERROR] There is no files in the input/metadata folder. Please drop some files in order to be analyzed.")
            exit()
        else:
            self.list_input_dir()

    #Function for listing the content of the input folder
    def list_input_dir(self):
        valid = False
        files_size = len(self.files)
        while not valid:
            print("\n"*3)
            print("What file do you want to analyze?")
            print("---------------------------------")
            for index,file in enumerate(self.files):
                print("{}). {}".format(index+1,file))
            try:
                self.option = int(input("Select an option: "))
            except:
                self.option = 0
            if self.option >= 1 and self.option <= files_size:
                valid = True
            else:
                if files_size == 1:
                    print("[ERROR] The option must be a numeric argument of 1")
                else:
                    print("[ERROR] The option must be a numeric argument from: 1 to {}".format(files_size))
        self.file_selected = [self.files[self.option-1]]

    #Function for maping the file selected to a folder format
    def __map_file_folder(self):
        self.file_selected = list(map(lambda x: "/home/saw/Scripts/input/metadata/{}".format(x),self.file_selected))

    #Function for calling the exiftool command
    def call(self):
        command = ["exiftool"]
        self.__map_file_folder()
        command.extend(self.file_selected)
        self.output = subprocess.Popen(command,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL).communicate()
        self.__parse()
        self.prettify(self.file_selected[0])

    #Function for calling the exiftool command with arguments
    def call_with_args(self,args,output=True,overwrite=False):
        command = ["exiftool"]
        #self.__map_file_folder()
        command.extend([args])

        #To overwrite original file
        if overwrite: command.extend(["-overwrite_original"])
        command.extend(self.file_selected)

        lines = subprocess.Popen(command,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL).communicate()[0].decode("utf-8").splitlines()
        if not output: lines = []
        response_command = {}
        for line in lines:
            #Cleaning useless characters from output
            response = self.key_value_regex.search(line).groupdict()
            key,value = response["key"].strip(),response["value"].strip()
            response_command[key] = value
        return response_command



    #Function for parsing the output of exiftool command
    def __parse(self):
        lines = list(map(lambda x: x.decode("iso-8859-1"),self.output[0].splitlines()))
        for line in lines:
            #Cleaning useless characters from output
            response = self.key_value_regex.search(line).groupdict()
            key,value = response["key"].strip(),response["value"].strip()
            self.metadata[key] = value

            #Detecting GPS Tags
            if self.location_regex.search(key):
                self.gps_tags[key] = value

    
    #Function for creating a regex which split key/value from output
    def __create_key_value_regex(self):
        self.key_value_regex = re.compile(r"(?P<key>[A-Za-z0-9 ]*):(?P<value>.*)")

    #Function for detecting Location tags in the output information
    def __detect_location_tags(self):
        self.location_regex = re.compile(r"\.*GPS\.*")

    #Function for extracting default latitude/longitude format
    def __extract_lat_long_values(self):
        self.lat_long_regex = re.compile(r"(?P<degree>[0-9\.]*) deg (?P<minutes>[0-9\.]*)' (?P<seconds>[0-9\.]*)\" (?P<orientation>[NSWE])")

    #Function for prettifying the output
    def prettify(self,file_name):
        print()
        print("--- OUTPUT from: {} ---".format(file_name))
        print()
        for key in self.metadata:
            print("* [{}]: {}".format(key,self.metadata[key]))


    #Function for converting latitude/longitude values into decimal number
    def lat_long_to_decimal(self,input_values):
        decimal_value = (1 if input_values["orientation"] in ["N","E"] else -1)*(float(input_values["degree"]) + float(input_values["minutes"])/60 + float(input_values["seconds"])/3600)
        return decimal_value

    #Function for parsing the reverse geocode output
    def __parse_place(self,place_tags):
        place_information = {}
        
        for tag in place_tags:
            if "street_number" in tag["types"]:
                place_information["street_number"] = tag["long_name"]
            elif "route" in tag["types"]:
                place_information["street_name"] = tag["long_name"]
            elif "locality" in tag["types"]:
                place_information["city_name"] = tag["long_name"]
            elif "administrative_area_level_2" in tag["types"]:
                place_information["province"] = tag["long_name"]
            elif "administrative_area_level_1" in tag["types"]:
                place_information["autonomous_community"] = tag["long_name"]
            elif "country" in tag["types"]:
                place_information["country"] = tag["long_name"]
            elif "postal_code" in tag["types"]:
                place_information["postal_code"] = tag["long_name"]
        
        return place_information

    #Function for collecting private information
    def collect_private_information(self):
        print("\n"*3)
        self.private_information = {}
        print("[{}] Collecting private information...".format(self.__get__time()))
        author = self.call_with_args("-Author")["Author"] if self.call_with_args("-Author") else ""
        if author != "": self.private_information["author"] = author

        camera_model_name = self.call_with_args("-Model")["Camera Model Name"] if self.call_with_args("-Model") else ""
        if camera_model_name != "": self.private_information["camera_model"] = camera_model_name

        software_version = self.call_with_args("-Software")["Software"] if self.call_with_args("-Software") else ""
        if software_version != "": self.private_information["software_version"] = software_version

        create_date = self.call_with_args("-CreateDate")["Create Date"] if self.call_with_args("-CreateDate") else ""
        if create_date != "": self.private_information["create_date"] = create_date

        maker = self.call_with_args("-Make")["Make"] if self.call_with_args("-Make") else ""
        if maker != "": self.private_information["maker"] = maker

        gps_altitude = self.call_with_args("-gpsaltitude")["GPS Altitude"] if self.call_with_args("-gpsaltitude") else ""
        if gps_altitude != "": self.private_information["gps_altitude"] = gps_altitude
        
        compression = self.call_with_args("-Compression")["Compression"] if self.call_with_args("-Compression") else ""
        if compression != "": self.private_information["compression"] = compression

        if self.gps_tags:
            self.private_information["gps_latitude"] = self.gps_tags["GPS Latitude"]
            self.private_information["gps_longitude"] = self.gps_tags["GPS Longitude"]

        if author != "" or camera_model_name != "" or software_version != "" or create_date != "" or maker != "" or gps_altitude != "" or compression != "":
            print("--- PRIVATE INFORMATION ---")
            for key in self.private_information:
                if key == "author": print("* Author: {}".format(self.private_information["author"]))
                elif key == "camera_model": print("* Camera Model: {}".format(self.private_information["camera_model"]))
                elif key == "software_version": print("* Software Version: {}".format(self.private_information["software_version"]))
                elif key == "create_date": print("* Create Date: {}".format(self.private_information["create_date"]))
                elif key == "maker": print("* Maker: {}".format(self.private_information["maker"]))
                elif key == "gps_altitude": print("* GPS Altitude: {}".format(self.private_information["gps_altitude"]))
                elif key == "gps_latitude": print("* GPS Latitude: {}".format(self.private_information["gps_latitude"]))
                elif key == "gps_longitude": print("* GPS Longitude: {}".format(self.private_information["gps_longitude"]))
                elif key == "compression": print("* Compression: {}".format(self.private_information["compression"]))
        else:
            print("[INFO] There is no additional private information to display.")



    #Function for displaying place informatino
    def __display_place(self,place_information):
        print("\n"*3)
        print("--- PLACE REVIEW ----")
        print("* Address: {}".format(place_information["street_name"]))
        print("* Address Number: {}".format(place_information["street_number"]))
        print("* Postal Code: {}".format(place_information["postal_code"]))
        print("* City: {}".format(place_information["city_name"]))
        print("* Province: {}".format(place_information["province"]))
        print("* Autonomous Community: {}".format(place_information["autonomous_community"]))
        print("* Country: {}".format(place_information["country"]))
        print("--- Coordinates ---")
        print("\t- Latitude: {}".format(place_information["coordinates"][0]))
        print("\t- Longitude: {}".format(place_information["coordinates"][1]))

    #Function for analyzing location tags
    def analyze_location_tags(self):
        print("\n"*7)

        valid = False
        while not valid:
            try:
                option = input("Do you want to analyze GPS location tags? [y/n]: ").lower()
                if option not in ["y","n"]:
                    print("[ERROR] The option must be: y or n [yes/no]")
                else: 
                    valid = True
            except:
                print("[ERROR] Incorrect option format. The selection must be: y or n [yes/no]")

        #Analize if option is enabled
        if option == "y":
            #Checking wheter or not the API is written in the proper configuration file
            self.__checking_api_key()
            gmaps = googlemaps.Client(key=self.api_key)
            latitude = 0
            Longitude = 0

            #Detecting latitude/longitude values
            for tag in self.gps_tags:
                if tag == "GPS Latitude":
                    latitude = self.lat_long_regex.search(self.gps_tags[tag]).groupdict()
                    latitude = self.lat_long_to_decimal(latitude)
                elif tag == "GPS Longitude":
                    longitude = self.lat_long_regex.search(self.gps_tags[tag]).groupdict()
                    longitude = self.lat_long_to_decimal(longitude)

            if latitude != 0 and longitude != 0:
                results = gmaps.reverse_geocode((latitude,longitude))
                for result in results:
                    if "street_address" in result["types"]:
                        place = result["address_components"]
                        break
                
                place_information = self.__parse_place(place)
                place_information["coordinates"] = (latitude,longitude)
                self.__display_place(place_information)

                #Ask the user for saving a screenshot
                self.__save_screen(place_information)
            else:
                print("[INFO] There is no GPS Location information available.")


    #Function for detecting the screen size of certain computer
    def __screen_size(self):
        response = subprocess.Popen(["xrandr"],stdout=subprocess.PIPE,stderr=subprocess.DEVNULL)
        screen_size = subprocess.Popen(['grep',"*"],stdin=response.stdout,stdout=subprocess.PIPE).communicate()
        screen_size = screen_size[0].decode("utf-8")
        screen_size = re.search(r"\s*([0-9x]*)\s*",screen_size).group(1).split("x")
        return screen_size

    #Function for saving an screenshot of the location map
    def __save_screen(self,place_information):
        print("\n"*3)
        valid = False
        while not valid:
            try:
                option = input("Do you want to save an screenshot of the map? [y/n]: ").lower()
                valid = True
            except:
                print("[ERROR] The option must be y or n [yes/no]")
        if option == "y":
            print("[{}] Saving screen in output/maps. Please wait a few seconds...".format(self.__get__time()))
            
            """
            CHROME_PATH = '/usr/bin/chromium-browser'
            CHROMEDRIVER_PATH = '/usr/bin/google-chrome'


            chrome_options = Options()  
            chrome_options.add_argument("--headless")  
            chrome_options.add_argument("--window-size={}".format(WINDOW_SIZE))
            chrome_options.binary_location = CHROME_PATH
            """
            """
            web_browser = webdriver.Chrome(executable_path=CHROMEDRIVER_PATH,chrome_options=chrome_options)             
            web_browser.get("https://www.google.es/maps/place/{},{}".format(place_information["coordinates"][0],place_information["coordinates"][1]))
            map_canvas = web_browser.find_element_by_class_name("widget-scene")
            location = map_canvas.location
            size = map_canvas.size
            """
            #The same name of the file being analyzed, but with png extension
            file_name = self.file_selected[0].rsplit("/",1)[1].split(".")[0]

            #Customized Window Size
            screen_size = self.__screen_size()
            WINDOW_SIZE = "{},{}".format(screen_size[0],screen_size[1])

            command = ["google-chrome",
                       "--headless",
                       "--disable-gpu",
                       "--window-size={}".format(WINDOW_SIZE),
                       "--screenshot=/home/saw/Scripts/output/maps/{}.png".format(file_name)
                       ,"https://www.google.es/maps/place/{},{}".format(place_information["coordinates"][0],place_information["coordinates"][1])]
            subprocess.Popen(command,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).communicate()                                              


            """
            web_browser.save_screenshot('output/maps/{}.png'.format(file_name)) 
            web_browser.quit()
            """
            
            
            im = Image.open('/home/saw/Scripts/output/maps/{}.png'.format(file_name)) 

            left = 430
            top = 0
            right = 430 + (int(screen_size[0]) - 430)
            bottom = top + int(screen_size[1])

            im = im.crop((left, top, right, bottom)) 
            im.save('/home/saw/Scripts/output/maps/{}.png'.format(file_name))
            print("[{}] Screenshot saved successfully.".format(self.__get__time()))
            im.show()

            #Dumping private information to a json file
            self.__dump_json()

    #Function for dumping all private information to a json file
    def __dump_json(self):
        print("\n"*3)
        print("[{}] Dumping private information to a json file (output/json)...".format(self.__get__time()))
        file_name = self.file_selected[0].split("/")[-1].split(".")[0]
        with open("/home/saw/Scripts/output/json/{}.json".format(file_name),"w+") as f:
            f.write(json.dumps(self.private_information))

    #Function for anonymizing a certain file
    def anonymize(self):
        print("[{}] Anonymizing: {}...".format(self.__get__time(),self.file_selected[0]))
        self.call_with_args("-all=",output=False,overwrite=True)
        print("[{}] Process complete successfully.".format(self.__get__time()))

    #Function for faking the metadata of certain file
    def fake(self):
        print("[{}] Starting faking procedure of file: {}...".format(self.__get__time(),self.file_selected[0]))
        fake_config = configparser.ConfigParser()
        fake_config.read("/home/saw/Scripts/conf/fake_options.cfg")
        fake_config.sections()

        #Manufacturer Name
        manufacturer_name = secrets.choice(["Apple","Google"])

        #Remove private metadata
        self.call_with_args("-all=",output=False,overwrite=True)

        #Change Author Name
        self.call_with_args("-Author='{}'".format(secrets.choice(json.loads(
            fake_config["OPTIONS"]["Author"]))),output=False,overwrite=True
        )

        #Change Camera Model
        self.call_with_args("-Model='{}'".format(secrets.choice(json.loads(
            fake_config["OPTIONS"]["CameraModelApple"] if manufacturer_name == "Apple" else fake_config["OPTIONS"]["CameraModelAndroid"]))),output=False,overwrite=True
        )

        #Change Software Version
        self.call_with_args("-Software='{}'".format(secrets.choice(json.loads(
            fake_config["OPTIONS"]["SoftwareApple"] if manufacturer_name == "Apple" else fake_config["OPTIONS"]["SoftwareAndroid"]))),output=False,overwrite=True
        )

        
        #Change Software Version
        self.call_with_args("-CreateDate='{}'".format((datetime.datetime.now() - datetime.timedelta(days=secrets.choice(range(1,100)))).strftime("%Y:%m:%d %H:%M:%S+01:00")),output=False,overwrite=True)
        
        #Change Maker
        self.call_with_args("-Make='{}'".format(secrets.choice(json.loads(
            fake_config["OPTIONS"]["MakerApple"] if manufacturer_name == "Apple" else fake_config["OPTIONS"]["MakerAndroid"]))),output=False,overwrite=True
        )
    
        #Compression
        self.call_with_args("-Compression='{}'".format(secrets.choice(json.loads(fake_config["OPTIONS"]["Compression"]))),output=False,overwrite=True)

        

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('-a','--anonymize', action='store_true',help='Anonymize certain file')
    parser.add_argument('-f','--fake',action='store_true',help='Fake private information available')
    parser.add_argument('-s','--show',action='store_true',help='Show metadata associated to a certain file')
    parser.add_argument('-v','--version', action='version',version='%(prog)s 1.0',help='Show current version of the program')

    results = parser.parse_args()

    if results.anonymize: #Anonymize File
        metadata_extractor = MetadataExtractor()
        metadata_extractor.introduction()
        metadata_extractor.check_input_dir()
        metadata_extractor.call()
        metadata_extractor.anonymize()
    elif results.fake: #Fake File
        metadata_extractor = MetadataExtractor()
        metadata_extractor.introduction()
        metadata_extractor.check_input_dir()
        metadata_extractor.call()
        metadata_extractor.fake()
    elif results.show:  #Show File
        metadata_extractor = MetadataExtractor()
        metadata_extractor.introduction()
        metadata_extractor.check_input_dir()
        metadata_extractor.call()
    else:   #No parameter execution
        metadata_extractor = MetadataExtractor()
        metadata_extractor.introduction()
        metadata_extractor.check_input_dir()
        metadata_extractor.call()
        metadata_extractor.collect_private_information()
        metadata_extractor.analyze_location_tags()
