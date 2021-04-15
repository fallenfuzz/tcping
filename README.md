# TCPING

A cross-platform ping program for ```TCP``` ports similar to Linux's ping utility. This program will send ```TCP``` probes to an ```IP address``` or a ```hostname``` specified by you and prints the results.

It uses different sequence numbering for successful and unsuccessful probes, so that when you look at the results after a while, and seeing for instance, a failed probe, understanding the total packet drops so far would be illustrative enough.

## Application

* Calculate packet loss.
* Assess latency of your network.
* Show min/avg/max probes latency.
* Monitor and audit your peers network.
* Calculate total up or downtime.
* Make sure a host is up in environments that ping is blocked.
* etc.

## Images

![WindowsVersion](/Images/windowsVersion.png)

## Demo

[![asciicast](https://asciinema.org/a/bNMtJKmujGEpfEhvDiTeSvtO4.svg)](https://asciinema.org/a/bNMtJKmujGEpfEhvDiTeSvtO4)

## Download for

* ### [Windows](https://github.com/pouriyajamshidi/tcping/releases/download/Windows-v1.0.0/tcping.exe)

* ### [Linux](https://github.com/pouriyajamshidi/tcping/releases/download/Linux-v1.0.0/tcping)

* ### [macOS](https://github.com/pouriyajamshidi/tcping/releases/download/macOS-v1.0.0/tcping)

## Usage

Go to the directory/folder in which you have downloaded the application.

### On ```Linux``` and ```macOS```:

```bash
sudo chmod +x tcping
```

For easier use, you can copy it to your system ```PATH``` like /bin/ or /usr/bin/

```bash
sudo cp tcping /bin/
```

Then run it like:

```bash
tcping www.example.com 443
```

OR

```bash
tcping 10.10.10.1 22
```

### On ```Windows```

I recommend ```Windows Terminal``` for the best experience and proper colorization.

For easier use, copy ```tcping.exe``` to your system ```PATH``` like C:\Windows\System32 or from your terminal application, go to the folder that contains the ```tcping.exe``` program.

Run it like:

```powershell
.\tcping www.example.com 443
```

OR

```powershell
tcping 10.10.10.1 22
```

**Please note, if you copy the program to your system ```PATH```, you don't need to specify ```.\``` to run the program anymore.**

## Tips

* While the program is running, upon pressing the ```enter``` key, the summary of all probes will be shown.

## Notes

This program is still in a ```beta``` stage. There are several shortcomings that I will rectify in the near future.

## Tested on

Windows, Linux and macOS

## Contributing

Pull requests are welcome.

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)