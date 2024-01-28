<p align="center"><a href="https://github.com/redskal/maul"><img alt="Maul Logo" src="assets/logo.png" width="50%" /></a></p>

[![Go](https://goreportcard.com/badge/github.com/redskal/maul)](https://goreportcard.com/report/github.com/redskal/maul) 

#### Overview
Maul is a simple tool which parses a list of URLs and splits them into paths (to a depth of two), subdomains, filenames, and parameter names. The idea came after watching a YT video ([NahamSec](https://twitter.com/NahamSec), I think) which talked about building custom wordlists containing realistic, up-to-date targets you're seeing in the wild.

I figured I'd do this by exporting URLs from Burp, processing them using the techniques in Maul, and build a master list and/or target-specific lists. The tool isn't perfect, but it's good enough to serve its purpose for now.

I suppose bug bounty hunters could also use this with a crawler and `anew` to get more granular updates about changes to anything in-scope for a program, too.

If you're exporting URLs post-engagement, you might consider removing anything that identifies clients from the files.

#### Install
Requirements:
- Go version 1.21+

To install just run the following command:

```bash
go install github.com/redskal/maul/cmd/maul@latest
```

#### Usage
It's pretty straightforward. Export a list of URLs from Burp, or pipe them in from something like `hakcrawler`.
```
Usage of maul:
-f string
                File to process.
-ef string
                Exclude files with given extensions. Comma-separated list. (default ".png,.jpg,.svg,.woff,.ttf,.eot")
-o string
                Directory to output files to. (default "./")
-t int
                Amount of threads to run. (default 50)

Input can also be supplied by piping it in.
Eg.
        $ cat urls.txt | maul
        $ maul < urls.txt

Output files are:
        files.txt      - any filenames found
        paths.txt      - any paths up to a depth of 2 (/path/here)
        subdomains.txt - any subdomains it can identify
        parameters.txt - names of any parameters it finds
```

#### To-dos
List of items to add or improve:
- Improve the file and parameter name parsing to avoid pulling in junk.
- Add blacklisting to avoid importing data that can identify clients.

#### License
Released under the [#YOLO Public License](https://github.com/YOLOSecFW/YoloSec-Framework/blob/master/YOLO%20Public%20License)