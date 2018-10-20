export CASC_APP=$1
export CASC_REGION="eu"
export CASC_CDN="eu"

if [ "$CASC_APP" == "w3" ]; then
	export CASC_DIR="/Applications/Warcraft III"
	export CASC_PATTERN="War3.mpq:Movies/*.avi"
elif [ "$CASC_APP" == "d3" ]; then
	export CASC_DIR="/Applications/Diablo III"
	export CASC_PATTERN="enUS/Data_D3/Locale/enUS/Cutscenes/Cinematic_1*.ogv"
elif [ "$CASC_APP" == "s1" ]; then
	export CASC_DIR="/Applications/StarCraft"
	export CASC_PATTERN="HD2/Smk/*.webm"
elif [ "$CASC_APP" == "pro" ]; then
	export CASC_DIR="/Applications/Overwatch"
	export CASC_PATTERN=""
elif [ "$CASC_APP" == "s2" ]; then
	export CASC_DIR="/Applications/StarCraft II"
	export CASC_PATTERN=""
fi
export CASC_PATTERN=""

make $2