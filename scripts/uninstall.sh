#!/usr/bin/env bash

BINARY_NAME="weasel"
INSTALL_PATH="/usr/local/bin"

echo "Uninstalling ${BINARY_NAME}..."

# check if the binary exists at the destination
if [ ! -f "${INSTALL_PATH}/${BINARY_NAME}" ]; then
	echo "Warning: ${BINARY_NAME} not found in ${INSTALL_PATH}. Nothing to uninstall."
	exit 0
fi

# confirm before uninstalling
read -p "Are you sure you want to uninstall ${BINARY_NAME} from ${INSTALL_PATH}? (y/n): " CONFIRM
if [[ "${CONFIRM}" != "y" && "${CONFIRM}" != "Y" ]]; then
	echo "Uninstallation cancelled."
	exit 0
fi

# remove the binary
sudo rm -f "${INSTALL_PATH}/${BINARY_NAME}" || {
	echo "Error: Failed to uninstall ${BINARY_NAME}. Do you have the necessary permissions?"
	exit 1
}

echo "${BINARY_NAME} has been successfully uninstalled from ${INSTALL_PATH}."
exit 0
