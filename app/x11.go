package app

import "fmt"

func buildX11UserCmds(de, username, distroID string) []string {
	deCmd := deStartCmd(de)
	if deCmd == "" {
		return nil
	}
	src := fmt.Sprintf(`. ~/.lazypoot/x11.sh`)
	rc := userRcFile(distroID)
	return []string{
		fmt.Sprintf("mkdir -p /home/%s/.lazypoot", username),
		fmt.Sprintf(`cat > /home/%s/.lazypoot/x11.sh << 'X11EOF'
#!/bin/bash
export DISPLAY=:0
%s &
X11EOF`, username, deCmd),
		fmt.Sprintf(`grep -q 'lazypoot/x11.sh' /home/%s/%s 2>/dev/null || echo '%s' >> /home/%s/%s`, username, rc, src, username, rc),
	}
}
