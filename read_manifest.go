package main

import (
	"bytes"
    "bufio"
	"fmt"
	"io/ioutil"
	"os"
	"encoding/xml"
	"strings"
	"./configparser"
  "./call_command"
  "regexp" 
)


//Contants for configuration options
const API_VERSION_FILENAME string = "/home/saw/Scripts/conf/api_version.cfg"
const PERMISSION_FILENAME string = "/home/saw/Scripts/conf/permission.cfg"
const NEW_LINE string = "\n"

//Terminal Colors
type Colors struct {
	Red string
	Green string
	Yellow string
	Blue string
	Magenta string
	Cyan string
	White string
	End string
}

//Constructor
func (c *Colors) New(){
   c.Red = "\x1b[91m"
   c.Green = "\x1b[92m"
   c.Yellow = "\x1b[93m"
   c.Blue = "\x1b[94m"
   c.Magenta = "\x1b[95m"
   c.Cyan = "\x1b[96m"
   c.White = "\x1b[97m"
   c.End = "\x1b[0m"
}

//Function for getting the color associated to certain grade
func (c Colors) GetGradeColor(grade string) string {
   var result string
   if grade == "DANGEROUS" {
	   result = c.Red
   }else if grade == "NEUTRAL" {
	   result = c.Blue
   }else if grade == "PROBLEMATIC" {
	   result = c.Yellow
   }else if grade == "GOOD" {
	   result = c.Green
   }
   return result
}

var colors Colors

//Manifest Overview
type ManifestLevel struct {
		Color Colors
		Manifest xml.Name `xml:"manifest"`
		Package_Name string `xml:"package,attr"`
		Schema string `xml:"android,attr"`
		Android_Version_Name string `xml:"versionName,attr"`
		Version_Code int8 `xml:"versionCode,attr"`
		Install_Location string `xml:"installLocation,attr"`
		App Application `xml:"application"`
		MinSDKVersion UseSDKTag `xml:"uses-sdk"`
		ScreenSupport ScreenSupportTag `xml:"supports-screens"`
		Permissions []PermissionTag `xml:"uses-permission"`
		PermissionsSdk23 []PermissionTag `xml:"uses-permission-sdk-23"`
		APIConfig map[string] string
		PermissionConfig map[string] string
	}

//Function for casting the manifest information to string
func (m ManifestLevel) String() string  {
	m.Color.New()
	var buffer bytes.Buffer
	s := fmt.Sprintf("%s%s%s\n",colors.Yellow,"------- Manifest Overview -------",colors.End)
	buffer.WriteString(s)
	s = fmt.Sprintf("* Package Name: %s\n",m.Package_Name)
	buffer.WriteString(s)
	s = fmt.Sprintf("* Schema: %s\n",m.Schema)
	buffer.WriteString(s)
	s = fmt.Sprintf("* Android Version Name: %s\n",m.Android_Version_Name)
	buffer.WriteString(s)
    if m.Version_Code != 0 {
        s = fmt.Sprintf("* Version Code: %d\n",m.Version_Code)
        buffer.WriteString(s)
    }
	s = fmt.Sprintf("* Install Location: %s\n",m.Install_Location)
	buffer.WriteString(s)
	buffer.WriteString(m.App.String())
	buffer.WriteString(m.MinSDKVersion.String(m.APIConfig))
	buffer.WriteString(m.ScreenSupport.String())
	buffer.WriteString(fmt.Sprintf("\n\t%s%s%s\n",colors.Yellow,"--- Permissions ---",colors.End))
     
    if len(m.PermissionsSdk23) != 0 {
           m.Permissions = append(m.Permissions,m.PermissionsSdk23...)
    }
    
    
	for _,value := range m.Permissions {
		buffer.WriteString(value.String())
		desc := strings.Split(m.GetPermissionDesc(value.Name),"|")
        
        if len(desc) != 2 {
                desc = []string{"UNKNOWN","Customized permission. Description not available."}
        }
		grade := desc[0]
		description := desc[1]
		color_permission := colors.GetGradeColor(grade)
		buffer.WriteString(fmt.Sprintf("%s\t\tͰ [%s] %s%s\n",color_permission,grade,description,colors.End))
		buffer.WriteString(NEW_LINE)
	}
	return buffer.String()
}

//Function for dumping Manifest Information to a json file
func (m ManifestLevel) ToJSON()string {
    var res []string
    res = append(res,fmt.Sprintf("\"package_name\": \"%s\"",m.Package_Name))
    res = append(res,fmt.Sprintf("\"schema\": \"%s\"",m.Schema))
    res = append(res,fmt.Sprintf("\"android_version\": \"%s\"",m.Android_Version_Name))
    if m.Version_Code != 0 {
        res = append(res,fmt.Sprintf("\"version_code\": \"%s\"",m.Version_Code))
    }else {
        res = append(res,fmt.Sprintf("\"version_code\": \"not specified\""))
    }
    res = append(res,fmt.Sprintf("\"install_location\": \"%s\"",m.Install_Location))

    var permission_value []string
    
    if len(m.PermissionsSdk23) != 0 {
           m.Permissions = append(m.Permissions,m.PermissionsSdk23...)
    }
        
    for _,value := range m.Permissions {
        //fmt.Println("ValueN: ",value.Name)
        desc := strings.Split(m.GetPermissionDesc(value.Name),"|")
        if len(desc) != 2 {
                desc = []string{"UNKNOWN","Customized permission. Description not available."}
        }
		grade := desc[0]
		description := desc[1]
		permission_value = append(permission_value,fmt.Sprintf("\"<%s>: %s -> %s\"",value.Name,description,grade))
    }
    
   permission_res := fmt.Sprintf("\"permissions\": [%s]",strings.Join(permission_value,","))

    
    res = append(res,permission_res)
 
    res = append(res,m.App.ToJSON())
    
    return fmt.Sprintf("{\"manifest\":{%s}}",strings.Join(res,", "))
}

//Function for writing the json file in a certain folder
func (m ManifestLevel) Write(name string,ext string) {
    if ext == "json"{
        f,_ := os.Create(fmt.Sprintf("/home/saw/Scripts/output/json/%s_manifest.json",strings.Split(name,".apk")[0]))

        defer f.Close()
        
        w := bufio.NewWriter(f)
        w.WriteString(m.ToJSON())
        w.Flush()
    }
}

//Function for setting the initial configuration to the program
func (m *ManifestLevel) SetConfig() {
    var api_config configparser.ConfigParser
    api_config.Load(API_VERSION_FILENAME)
    var permission_config configparser.ConfigParser
    permission_config.Load(PERMISSION_FILENAME)
    m.APIConfig = api_config.GetConfigOpts()
    m.PermissionConfig = permission_config.GetConfigOpts()
}

//Function for getting the description information for a certain permission
func (m ManifestLevel) GetPermissionDesc(permission string) string {
    index_dot := strings.LastIndex(permission,".") + 1
	permission = permission[index_dot:]
	return m.PermissionConfig[permission]
}


//Application Section
type Application struct {
	ApplicationTag xml.Name `xml:"application"`
	Icon string `xml:"icon,attr"`
	Name string `xml:"name,attr"`
	Label string `xml:"label,attr"`
    Metadata []MetadataTag `xml:"meta-data"`
	Activity []ActivityTag `xml:"activity"`
	Service []ServiceTag `xml:"service"`
	Receiver []ReceiverTag `xml:"receiver"`
    Provider []ProviderTag `xml:"provider"`
}

//Function for casting the Application information to string
func (a Application) String() string {
	var buffer bytes.Buffer
	s := fmt.Sprintf("\n\t%s%s%s\n",colors.Yellow,"--- Application ---",colors.End)
	buffer.WriteString(s)
	s = fmt.Sprintf("\t* Icon: %s\n",a.Icon)
	buffer.WriteString(s)
	s = fmt.Sprintf("\t* Name: %s\n",a.Name)
	buffer.WriteString(s)
	s = fmt.Sprintf("\t* Label: %s\n",a.Label)
	buffer.WriteString(s)
    
    if len(a.Metadata) != 0 {
        s = fmt.Sprintf("\n\t\t\t%s%s%s\n",colors.Yellow,"--- Metadata ---",colors.End)
        buffer.WriteString(s)
    }
	for _,value := range a.Metadata {
		s = fmt.Sprintf(value.String())
		buffer.WriteString(s)
	}
    
    if len(a.Activity) > 0 {
        buffer.WriteString(fmt.Sprintf("%s\n\n\t\t--- ACTIVITIES ---\n\n%s",colors.Red,colors.End))
    }
	for _,value := range a.Activity{
		buffer.WriteString(value.String())
	}
    if len(a.Service) > 0 {
        buffer.WriteString(fmt.Sprintf("%s\n\n\t\t--- SERVICIES ---\n\n%s",colors.Red,colors.End))
    }
	for _,value := range a.Service{
		buffer.WriteString(value.String())
	}
	
    if len(a.Receiver) > 0 {
        buffer.WriteString(fmt.Sprintf("%s\n\n\t\t--- RECEIVERS ---\n\n%s",colors.Red,colors.End))
    }
	for _,value := range a.Receiver{
		buffer.WriteString(value.String())
	}
	
    if len(a.Provider) > 0 {
        buffer.WriteString(fmt.Sprintf("%s\n\n\t\t--- PROVIDERS ---\n\n%s",colors.Red,colors.End))
    }
	for _,value := range a.Provider{
		buffer.WriteString(value.String())
	}
	
	return buffer.String()
}

//Function for writing the Application Information to a json file
func (a Application) ToJSON() string {
    var res []string
    res = append(res,fmt.Sprintf("\"icon\":\"%s\"",a.Icon))
    res = append(res,fmt.Sprintf("\"name\":\"%s\"",a.Name))
    res = append(res,fmt.Sprintf("\"label\":\"%s\"",a.Label))
   
    var metadata_value []string
    
    for _,value := range a.Metadata {
        metadata_value = append(metadata_value,value.ToJSON())
    }
    
    metadata_res := fmt.Sprintf("\"metadata\": [%s]",strings.Join(metadata_value,",\n"))

    res = append(res,metadata_res)
    
    var activity_value []string
    
    for _,value := range a.Activity {
        activity_value = append(activity_value,value.ToJSON())
    }
    
    activity_res := fmt.Sprintf("\"activities\": [%s]",strings.Join(activity_value,",\n"))
    
    res = append(res,activity_res)
    	
    var service_value []string
    
    for _,value := range a.Service {
        service_value = append(service_value,value.ToJSON())
    }
    
    service_res := fmt.Sprintf("\"services\": [%s]",strings.Join(service_value,","))
		
    res = append(res,service_res)
    
    var receiver_value []string
    
    for _,value := range a.Receiver {
        receiver_value = append(receiver_value,value.ToJSON())
    }
    
    receiver_res := fmt.Sprintf("\"receivers\": [%s]",strings.Join(receiver_value,",\n"))
	
    res = append(res,receiver_res)
    
    var provider_value []string
    
    for _,value := range a.Provider {
        provider_value = append(provider_value,value.ToJSON())
    }
    
    provider_res := fmt.Sprintf("\"providers\": [%s]",strings.Join(provider_value,",\n"))
    
    res = append(res,provider_res)
    
    return fmt.Sprintf("\"application\":{%s}",strings.Join(res,",\n"))
}


//Service Sections
type ServiceTag struct {
    Exported string `xml:"exported,attr"`
    Name string `xml:"name,attr"`
    Permission string `xml:"permission,attr"`
    Intent IntentTag `xml:"intent-filter"`
}

//Function for casting the Service Information into a string
func (sv ServiceTag) String() string {
    var buffer bytes.Buffer
    
	if sv.Exported != "" {
		s := fmt.Sprintf("\t\t* Exported: %s\n",sv.Exported)
		buffer.WriteString(s)
	}

	if sv.Name != "" {
		s := fmt.Sprintf("\t\t* Name: %s\n",sv.Name)
		buffer.WriteString(s)
	}

	if sv.Permission != "" {
		s := fmt.Sprintf("\t\t* Permission: %s\n",sv.Permission)
		buffer.WriteString(s)
	}
	
	if sv.Intent.String() != "" {
        buffer.WriteString(sv.Intent.String())
    }
    
	//Service Empty?
	if len(buffer.String()) == 0{
        return ""
    }
	return fmt.Sprintf("\n\t\t%s%s%s\n",colors.Yellow,"--- Service  ---",colors.End) + buffer.String()
}

//Function for dumping the Service Information to a json file
func (sv ServiceTag) ToJSON() string {
    var res []string
    if sv.Exported != "" {
    res = append(res,fmt.Sprintf("\"exported\":\"%s\"",sv.Exported))
    }
    if sv.Name != "" {
    res = append(res,fmt.Sprintf("\"name\":\"%s\"",sv.Name))
   }
   if sv.Permission != "" {
    res = append(res,fmt.Sprintf("\"permission\":\"%s\"",sv.Permission))
   }
   
   if sv.Intent.String() != "" {
    res = append(res,fmt.Sprintf("\"intent\": %s",sv.Intent.ToJSON()))
   }
   s := fmt.Sprintf("{%s}",strings.Join(res,","))
   return s
}


//Metadata Sections
type MetadataTag struct {
	Name string `xml:"name,attr"`
	Value string `xml:"value,attr"`
	Resource string `xml:"resource,attr"`
}

//Function for casting the Metadata Information into a string
func (m MetadataTag) String() string {
   var buffer bytes.Buffer
   s := fmt.Sprintf("\t\t\t* Name: %s\n",m.Name)
   buffer.WriteString(s)
   if m.Value != "" {
   	s = fmt.Sprintf("\t\t\t* Value: %s\n",m.Value)
 	buffer.WriteString(s)
   }
   if m.Resource != "" {
   	s = fmt.Sprintf("\t\t\t* Resource: %s\n",m.Resource)
   	buffer.WriteString(s)
   }
   buffer.WriteString(NEW_LINE)
   return buffer.String()
}

//Function for writing the Metadata Information to a json file
func (m MetadataTag) ToJSON() string {
    var res []string
    res = append(res,fmt.Sprintf("\"name\":\"%s\"",m.Name))
   if m.Value != "" {
    res = append(res,fmt.Sprintf("\"value\":\"%s\"",m.Value))
   }
   if m.Resource != "" {
    res = append(res,fmt.Sprintf("\"resource\":\"%s\"",m.Resource))
   }
   s := fmt.Sprintf("{%s}",strings.Join(res,","))
   return s
}

//Uses-SDK
type UseSDKTag struct {
	MinSDKVersion string `xml:"minSdkVersion,attr"`
}

//Function for getting the minimum SDK needed
func (u UseSDKTag) GetMinSDK() string {
	return u.MinSDKVersion
}

//Function for mapping the SDK version to Android version
func (u *UseSDKTag) String(c map[string] string) string{
	var buffer bytes.Buffer
	s := fmt.Sprintf("\n\t%s%s%s\n",colors.Yellow,"--- SDK Used ---",colors.End)
	buffer.WriteString(s)
    if c[u.MinSDKVersion] == "" {
        s = fmt.Sprintf("\t* Minimum SDK Version: not specified.\n")
    }else{
        s = fmt.Sprintf("\t* Minimum SDK Version: %s\n",c[u.MinSDKVersion])
    }
    buffer.WriteString(s)
	return buffer.String()
}

//Function for writing the SDK version into a json file
func (u UseSDKTag) ToJSON(c map[string] string) string {
        s := fmt.Sprintf("{\"min_sdk\": \"%s\"}",c[u.MinSDKVersion])
        return s
}

//Action Section
type ActionSection struct {
	Name string `xml:"name,attr"`
}

//Category Section
type CategorySection struct {
	Name string `xml:"name,attr"`
}

//Data Section
type DataSection struct {
	MimeType string `xml:"mimeType,attr"`
}

//Intent Section
type IntentTag struct {
	Action []ActionSection `xml:"action"`
	Category CategorySection `xml:"category"`
	Data DataSection `xml:"data"`
}

//Function for casting Intent Information into string
func (i IntentTag) String() string {
	var buffer bytes.Buffer
    
    for _,value := range i.Action {
        s := fmt.Sprintf("\t\t\t* Action: %s\n",value.Name)
		buffer.WriteString(s)    
    }
    

	if i.Category.Name != "" {
		s := fmt.Sprintf("\t\t\t* Category: %s\n",i.Category.Name)
		buffer.WriteString(s)
	}

	if i.Data.MimeType != "" {
		s := fmt.Sprintf("\t\t\t* Data: %s\n",i.Data.MimeType)
		buffer.WriteString(s)
	}
	
	//Intent Empty?
	if len(buffer.String()) == 0{
        return ""
    }
	return fmt.Sprintf("\n\t\t\t%s%s%s\n",colors.Yellow,"--- Intent  ---",colors.End) + buffer.String()
}

//Function for dumping Intent Information to a json file
func (i IntentTag) ToJSON() string {
    var res []string
	for _,value := range i.Action {
		res = append(res,fmt.Sprintf("\"action\": \"%s\"",value.Name))
	}

	if i.Category.Name != "" {
        res = append(res,fmt.Sprintf("\"category\": \"%s\"",i.Category.Name))
	}

	if i.Data.MimeType != "" {
        res = append(res,fmt.Sprintf("\"data\": \"%s\"",i.Data.MimeType))
	}
    s:= fmt.Sprintf("{%s}",strings.Join(res,","))
    if s != "{}"{
        return s
    }else {
        return ""
    }
}


//Activity Section
type ActivityTag struct {
	ConfigChanges string `xml:"configChanges,attr"`
	Name string `xml:"name,attr"`
	Label string `xml:"label,attr"`
	ScreenOrientation string `xml:"screenOrientation,attr"`
	Intent IntentTag `xml:"intent-filter"`
	Metadata []MetadataTag `xml:"meta-data"`
}

//Function for casting Activity Information into string
func (a ActivityTag) String() string {
	var buffer bytes.Buffer
	s := fmt.Sprintf("\n\t\t%s%s%s\n",colors.Yellow,"--- Activity ---",colors.End)
	buffer.WriteString(s)
	s = fmt.Sprintf("\t\t* Config Changes: %s\n",a.ConfigChanges)
	sx := strings.Replace(s,"|" ,", ",-1)
	buffer.WriteString(sx)
	s = fmt.Sprintf("\t\t* Name: %s\n",a.Name)
	buffer.WriteString(s)
    if a.Label != "" {
        s = fmt.Sprintf("\t\t* Label: %s\n",a.Label)
    }else {
        s = fmt.Sprintf("\t\t* Label: -\n")
    }
	buffer.WriteString(s)
    if a.ScreenOrientation != "" {
        s = fmt.Sprintf("\t\t* Screen Orientation: %s\n",a.ScreenOrientation)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Screen Orientation: -\n")
        buffer.WriteString(s)
    }
    
 	buffer.WriteString(a.Intent.String())
    if len(a.Metadata) != 0 {
        s = fmt.Sprintf("\n\t\t\t%s%s%s\n",colors.Yellow,"--- Metadata ---",colors.End)
        buffer.WriteString(s)
    }
	for _,value := range a.Metadata {
		s = fmt.Sprintf(value.String())
		buffer.WriteString(s)
	}
	return buffer.String()
}

//Function for dumping Activity Information to json file
func (a ActivityTag) ToJSON() string {
    var res []string
    res = append(res,strings.Replace(fmt.Sprintf("\"config_changes\": \"%s\"",a.ConfigChanges),"|",", ",-1))
    res = append(res,fmt.Sprintf("\"name\": \"%s\"",a.Name))
    res = append(res,fmt.Sprintf("\"label\": \"%s\"",a.Label))
    res = append(res,fmt.Sprintf("\"screen_orientation\": \"%s\"",a.ScreenOrientation))
    if a.Intent.ToJSON() != "" {
        res = append(res,fmt.Sprintf("\"intent\": %s",a.Intent.ToJSON()))
    }
    
    if len(a.Metadata) != 0{
    var metadata []string
    for _,value := range a.Metadata {
        metadata = append(metadata,value.ToJSON())
    }
    
    value_res := strings.Join(metadata,",")
    value_res = "\"metadata\": " + value_res
    
    res = append(res,value_res)
    }    
    
    return fmt.Sprintf("{%s}",strings.Join(res,","))
}

//Screen Support Tag
type ScreenSupportTag struct {
	LargeScreens bool `xml:"largeScreens,attr"`
	NormalScreens bool `xml:"normalScreens,attr"`
	SmallScreens bool `xml:"smallScreens,attr"`
	AnyDensity bool `xml:"anyDensity,attr"`
}

//Function for casting Screen Support Information into a string
func (st ScreenSupportTag) String() string {
	var buffer bytes.Buffer
	s := fmt.Sprintf("\n\t%s%s%s\n",colors.Yellow,"--- Screen Support ---",colors.End)
	buffer.WriteString(s)
	if st.LargeScreens == true {
		s = "\t* Large Screens: Yes\n"
	} else{
		s = "\t* Large Screens: No\n"
	}
	buffer.WriteString(s)
	if st.NormalScreens == true{
		s = "\t* Normal Screens: Yes\n"
	} else{
		s = "\t* Normal Screens: No\n"
	}
	buffer.WriteString(s)
	if st.SmallScreens == true{
		s = "\t* Small Screens: Yes\n"
	} else{
		s = "\t* Small Screens: No\n"
	}
	buffer.WriteString(s)
	if st.AnyDensity == true{
		s = "\t* Any Density?: Yes\n"
	} else {
		s = "\t* Any Density?: No\n"
	}
	buffer.WriteString(s)
	return buffer.String()
}

//Function for writing Screen Support Information to a json file
func (st ScreenSupportTag) ToJSON() string {
    res := make([]string, 1)
    if st.LargeScreens == true {
        res = append(res,"\"large_screens\": true")
	} else{
		res = append(res,"\"large_screens\": false")
	}
	
    if st.NormalScreens == true{
		res = append(res,"\"normal_screens\": true")
	} else{
        res = append(res,"\"normal_screens\": false")
	}
	
    if st.SmallScreens == true{
		res = append(res,"\"small_screens\": true")
	} else{
		res = append(res,"\"small_screens\": false")
	}
	
    if st.AnyDensity == true{
		res = append(res,"\"any_density\":true")
	} else {
        res = append(res,"\"any_density\": false")
    }
    s:= fmt.Sprintf("{%s}",strings.Join(res,","))
    return s
}

//Permission Information
type PermissionTag struct {
	Name string `xml:"name,attr"`
}


//Function for casting Permission Information to a string
func (p PermissionTag) String() string {
	var buffer bytes.Buffer
	s := fmt.Sprintf("\t* Name: %s\n",p.Name)
	buffer.WriteString(s)
	return buffer.String()
}

//Function for writing Permission Information to a json file
func (p PermissionTag) ToJSON() string {
   s:= fmt.Sprintf("{\"name\": %s}",p.Name)
   return s   
}


//Receiver Section
type ReceiverTag struct {
	Enabled string `xml:"enabled,attr"`
	Exported string `xml:"exported,attr"`
	Name string `xml:"name,attr"`
	Intent IntentTag `xml:"intent-filter"`
}

//Function for casting Receiver Information to a string
func (r ReceiverTag) String() string {
	var buffer bytes.Buffer
	s := fmt.Sprintf("\n\t\t%s%s%s\n",colors.Yellow,"--- Receiver ---",colors.End)
    buffer.WriteString(s)
    
    if r.Enabled != "" {
        s = fmt.Sprintf("\t\t* Enabled: %s\n",r.Enabled)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Enabled: -\n")
        buffer.WriteString(s)
    }
    if r.Exported != "" {
        s = fmt.Sprintf("\t\t* Exported: %s\n",r.Exported)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Exported: -\n")
        buffer.WriteString(s)
    }
    
    if r.Name != "" {
        s = fmt.Sprintf("\t\t* Name: %s\n",r.Name)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Name: -\n")
        buffer.WriteString(s)
    }
    
    if r.Intent.String() != "" {
        buffer.WriteString(r.Intent.String())
    }
    
	return buffer.String()
}

//Function for dumping Receiver Information to a json file
func (r ReceiverTag) ToJSON() string {
    var res []string
    res = append(res,strings.Replace(fmt.Sprintf("\"enabled\": \"%s\"",r.Enabled),"|",", ",-1))
    res = append(res,fmt.Sprintf("\"exported\": \"%s\"",r.Exported))
    res = append(res,fmt.Sprintf("\"name\": \"%s\"",r.Name))
    res = append(res,fmt.Sprintf("\"intent\": %s",r.Intent.ToJSON()))
    return fmt.Sprintf("{%s}",strings.Join(res,","))
}


//Provider Section
type ProviderTag struct {
	Authorities string `xml:"authorities,attr"`
	Exported string `xml:"exported,attr"`
	GrantUriPermissions string `xml:"grantUriPermissions,attr"`
    Name string `xml:"name,attr"`
	Metadata []MetadataTag `xml:"meta-data"`
}

//Function for casting Provider Information into a string
func (p ProviderTag) String() string {
	var buffer bytes.Buffer
	s := fmt.Sprintf("\n\t\t%s%s%s\n",colors.Yellow,"--- Provider ---",colors.End)
    buffer.WriteString(s)
    
    if p.Authorities != "" {
        s = fmt.Sprintf("\t\t* Authorities: %s\n",p.Authorities)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Authorities: -\n")
        buffer.WriteString(s)
    }
    if p.Exported != "" {
        s = fmt.Sprintf("\t\t* Exported: %s\n",p.Exported)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Exported: -\n")
        buffer.WriteString(s)
    }
    
    if p.GrantUriPermissions != "" {
        s = fmt.Sprintf("\t\t* Grant URI Permissions: %s\n",p.GrantUriPermissions)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Grant URI Permissions: -\n")
        buffer.WriteString(s)
    }
    
    if p.Name != "" {
        s = fmt.Sprintf("\t\t* Name: %s\n",p.Name)
        buffer.WriteString(s)
    }else {
        s = fmt.Sprintf("\t\t* Name: -\n")
        buffer.WriteString(s)
    }
    
    if len(p.Metadata) != 0 {
        s = fmt.Sprintf("\n\t\t\t%s%s%s\n",colors.Yellow,"--- Metadata ---",colors.End)
        buffer.WriteString(s)
    }
	for _,value := range p.Metadata {
		s = fmt.Sprintf(value.String())
		buffer.WriteString(s)
	}
    
	return buffer.String()
}

//Function for dumping Provider Information into a json file
func (p ProviderTag) ToJSON() string {
    var res []string
    res = append(res,fmt.Sprintf("\"authorities\": \"%s\"",p.Authorities))
    res = append(res,fmt.Sprintf("\"exported\": \"%s\"",p.Exported))
    res = append(res,fmt.Sprintf("\"grant_uri_permissions\": \"%s\"",p.GrantUriPermissions))
    res = append(res,fmt.Sprintf("\"name\": \"%s\"",p.Name))

    var metadata_value []string
    
    for _,value := range p.Metadata {
        metadata_value = append(metadata_value,value.ToJSON())
    }
    
    if len(metadata_value) > 0 {
        metadata_res := fmt.Sprintf("\"metadata\": [%s]",strings.Join(metadata_value,",\n"))
        res = append(res,metadata_res)
    }
    
    return fmt.Sprintf("{%s}",strings.Join(res,","))
}

//Function for checking into input folder, in order to list the number of available apps
func Check_Input_Folder() string {
    c:= call_command.New("ls")
    c.AddArgs([]string {"/home/saw/Scripts/input/"})
    return c.Call() 
}

//Function for displaying the program name
func Introduction(c Colors) {
 fmt.Println(c.Red,`
███╗   ███╗ █████╗ ███╗   ██╗██╗███████╗███████╗███████╗████████╗    ██╗███╗   ██╗████████╗███████╗██████╗ ██████╗ ██████╗ ███████╗████████╗███████╗██████╗ 
████╗ ████║██╔══██╗████╗  ██║██║██╔════╝██╔════╝██╔════╝╚══██╔══╝    ██║████╗  ██║╚══██╔══╝██╔════╝██╔══██╗██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝██╔══██╗
██╔████╔██║███████║██╔██╗ ██║██║█████╗  █████╗  ███████╗   ██║       ██║██╔██╗ ██║   ██║   █████╗  ██████╔╝██████╔╝██████╔╝█████╗     ██║   █████╗  ██████╔╝
██║╚██╔╝██║██╔══██║██║╚██╗██║██║██╔══╝  ██╔══╝  ╚════██║   ██║       ██║██║╚██╗██║   ██║   ██╔══╝  ██╔══██╗██╔═══╝ ██╔══██╗██╔══╝     ██║   ██╔══╝  ██╔══██╗
██║ ╚═╝ ██║██║  ██║██║ ╚████║██║██║     ███████╗███████║   ██║       ██║██║ ╚████║   ██║   ███████╗██║  ██║██║     ██║  ██║███████╗   ██║   ███████╗██║  ██║
╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝╚═╝     ╚══════╝╚══════╝   ╚═╝       ╚═╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
 `,c.End)
}
	
//Function for selecting an apk file among the applications available
func Select_APK(c Colors,response string) string {
     var valid bool = false 
     var apk_selected string
     re := regexp.MustCompile(".*\\.apk$")

     for !valid {
     apk_apps := strings.Split(response,"\n")
     n_apks := len(apk_apps)
     apks_aux := apk_apps[:n_apks-1]

     //Remove not apk files
    var apks []string
    for _,apk := range(apks_aux){
      if re.FindString(apk) != "" {
        apks = append(apks,apk)
      }
    }



     n_apks -= 1
     fmt.Println("Lista de apks, encontradas en el directorio 'input': ")
     fmt.Println("---------------------------------------------------- ")
     for index,apk := range apks {
            fmt.Printf("%d) %s\n",index+1,apk)   
     }

     var option int
     fmt.Println("Selecciona la app deseada, indicando su número: ")
     fmt.Scanf("%d",&option)
     switch {
         case option >=1 && option <= n_apks: valid = true
         default: valid = false
     }
     if !valid {
         fmt.Printf("%s[ERROR] No es una opción correcta. Introduzca un número entre: 1 y %d%s\n\n\n",c.Red,n_apks,c.End)
     }else{        
        fmt.Println("APK Seleccionada: ")
        for index,apk := range apks {
            if index+1 == option {
                    fmt.Printf("%s[*] %s\n%s",c.Green,apk,c.End)
                    apk_selected = apk
            }else {
                    fmt.Printf("[ ] %s\n",apk)
            }
        }
        Run_Decompile(c,apk_selected)
     }
     }
   return apk_selected
}

//Function for running the decompile process of certain Android app
func Run_Decompile(color Colors,apk string) {
    fmt.Println("Comenzando el proceso de decompilación...")
    c:= call_command.New("/usr/local/bin/apktool")
    c.AddArgs([]string {"d","-f","/home/saw/Scripts/input/"+apk,"-o","/home/saw/Scripts/output/"+apk})
    
    res := c.Call()
    if res != "" {
     fmt.Println(color.Green,"Proceso de Decompilación realizado con éxito...",color.End)   
    }else {
     fmt.Println(color.Red,"[ERROR] No se ha podido decompilar la aplicación de Android...",color.End)
    os.Exit(-2)
    }
}

//Main Function
func main(){
    colors.New()
    Introduction(colors)
    var apk_selected string
    
    var checking_response string = Check_Input_Folder()
    
    if checking_response == "" {
        fmt.Println("[ERROR] El directorio 'input' no incluye ninguna app. Por favor, introduzca alguna para continuar.")
        os.Exit(-1)
    }else
    {
    	//Selecting an Android app
        apk_selected = Select_APK(colors,checking_response)
    }
    
    manifest_path := fmt.Sprintf("/home/saw/Scripts/output/%s/AndroidManifest.xml",apk_selected)
	xmlFile,err := os.Open(manifest_path)
	if err != nil {
		fmt.Println("[ERROR] Al abrir el manifiesto.")
	}
	
	//Analyzing XML format
	x,_ := ioutil.ReadAll(xmlFile)
	var m ManifestLevel

	m.SetConfig()

	//Unmarshalling procedure
	xml.Unmarshal(x,&m)

	//Writing results to json and displaying the analysis on screen
    m.Write(apk_selected,"json")
    fmt.Println(m)
}
