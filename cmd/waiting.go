package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/3timeslazy/nix-search-tv/config"
	"github.com/3timeslazy/nix-search-tv/style"
)

const waitingMessage = "Indexing packages..."

const nixLogo = `
             .***        :::::.      :::
            -*****        ::::::    :::::.
             ******        ::::::  ::::::
              +*****        ::::::::::::
               =*****        ::::::::::
        ********************* :::::::.       +
       ***********************:::::::       ***
      -------------------------.::::::     *****
               ::::::            .:::::   ******
              :::::.              .::::: ******-
             :::::.                .::: ******
  :::::::::::::::                    : *************
 :::::::::::::::                      ***************
  ::::::::::::: *.                   ***************
        :::::: ***:                :*****
       :::::: *****=              -*****
      ::::::   *****+            +*****
      :::::     ****** .........................
       :::       ******.:::::::::::::::::::::::
        :       :******* :::::::::::::::::::::
               =*********       .::::::
              +***********        ::::::
             ******  +*****        ::::::
            .*****    =*****        :::::.
              ***      -*****        :::.
`

func PrintWaiting(out io.Writer) {
	out.Write(append([]byte(waitingMessage), []byte("\n")...))
}

func PreviewWaiting(out io.Writer, conf config.Config) {
	s := []string{
		nixLogo,
		"Looking for packages updates... It shouldn't take more than a few seconds.",
		"Next time, this message won't be here until the next indexing",
		"",
		"Indexing happens in two cases:",
		"  - It's the first run of the program",
		fmt.Sprintf("  - It's been more than %s since the last indexing", time.Duration(conf.UpdateInterval).String()),
		"",
		fmt.Sprintf("Btw (I use nix), you can set how often to update the index with %s option", style.TextStyle.Bold(config.UpdateIntervalTag)),
		fmt.Sprintf("Also, you can disable this message by setting %s to false", style.TextStyle.Bold(config.EnableWaitingMessageTag)),
		"",
		"Issues, comments and any contributions are welcome here:",
		"  https://github.com/3timeslazy/nix-search-tv",
		"",
		"Thank you for using nix-search-tv!",
	}

	msg := strings.Join(s, "\n")
	out.Write([]byte(msg))
}
