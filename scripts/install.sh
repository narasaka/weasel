#!/usr/bin/env bash

BINARY_NAME="weasel"
INSTALL_PATH="/usr/local/bin"
SOURCE_PATH="./${BINARY_NAME}"

echo "Installing ${BINARY_NAME}..."

# check if the binary exists
if [ ! -f "${SOURCE_PATH}" ]; then
	echo "Error: Binary ${SOURCE_PATH} not found. Please build the project first."
	exit 1
fi

# check if a file with the same name already exists at the destination
if [ -f "${INSTALL_PATH}/${BINARY_NAME}" ]; then
	read -p "A file named ${BINARY_NAME} already exists in ${INSTALL_PATH}. Overwrite? (y/n): " OVERWRITE
	if [[ "${OVERWRITE}" != "y" && "${OVERWRITE}" != "Y" ]]; then
		echo "Installation cancelled."
		exit 0
	fi
fi

# install the binary with executable permissions
sudo install -m 755 "${SOURCE_PATH}" "${INSTALL_PATH}/" || {
	echo "Error: Failed to install ${BINARY_NAME}. Do you have the necessary permissions?"
	exit 1
}

echo "${BINARY_NAME} has been successfully installed to ${INSTALL_PATH}."
exit 0
