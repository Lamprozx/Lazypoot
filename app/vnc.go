package app

import "fmt"

func deStartCmd(de string) string {
	switch de {
	case "xfce":
		return "startxfce4"
	case "lxqt":
		return "startlxqt"
	case "openbox":
		return "openbox-session"
	case "gnome":
		return "gnome-session"
	case "kde":
		return "startplasma-x11"
	case "awesome":
		return "awesome"
	case "i3":
		return "i3"
	}
	return ""
}

func buildVncUserCmds(vncPassword, username, distroID string) []string {
	src := fmt.Sprintf(`. ~/.lazypoot/vnc.sh`)
	rc := userRcFile(distroID)
	return []string{
		fmt.Sprintf("mkdir -p /home/%s/.lazypoot", username),
		fmt.Sprintf(`cat > /home/%s/.lazypoot/vnc.sh << 'VNCEOF'
#!/bin/bash
DISPLAY_NUM=1
GEOMETRY="1280x720"
PASSWORD="%s"
if ! vncserver -list 2>/dev/null | grep -q ":$DISPLAY_NUM"; then
	mkdir -p ~/.vnc
	echo "$PASSWORD" | vncpasswd -f > ~/.vnc/passwd 2>/dev/null
	chmod 600 ~/.vnc/passwd 2>/dev/null
	vncserver :$DISPLAY_NUM -geometry $GEOMETRY
fi
VNCEOF`, username, vncPassword),
		fmt.Sprintf(`grep -q 'lazypoot/vnc.sh' /home/%s/%s 2>/dev/null || echo '%s' >> /home/%s/%s`, username, rc, src, username, rc),
	}
}
