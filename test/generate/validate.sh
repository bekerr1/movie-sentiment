
SCRIPT_DIR=$(dirname "$0")
case "$SCRIPT_DIR" in
  .)  SCRIPT_DIR=$(pwd) ;;
  /*) ;;
  *)  SCRIPT_DIR=$(pwd)/"$SCRIPT_DIR" ;;
esac

count=0
for phrase in "Waste Of time" "Boring" "Terrible acting" "Predictable" "Hated it" "Too long" "Bad movie" "Disappointing" "Overrated" "Awful"; do 
	countThisPhrase=$(cat $SCRIPT_DIR/aggregate_sentiment.txt | grep -i "$phrase" | wc -l)
	count=$((count + countThisPhrase))
done

echo "Should be $count negative phrases overall"
