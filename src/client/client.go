package client

const BinaryName = "caspbx-cli"

func UserAgent(version string) string {
	return BinaryName + "/" + version
}
