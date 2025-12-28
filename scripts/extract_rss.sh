#!/bin/bash

# List of blogs from blogs.txt
blogs=(
    "https://danielmiessler.com"
    "https://docs.unsloth.ai"
    "https://www.anthropic.com/engineering"
    "https://eugenneyan.com"
    "https://blog.ezyang.com"
    "https://ampcode.com/news"
    "https://abseil.io"
    "https://karpathy.bearblog.dev"
    "https://lucumr.pocoo.org"
    "https://mariozechner.net"
    "https://ngrok.com/blog"
    "https://nicolaygerold.com"
    "https://rlancemartin.github.io"
    "https://simonwillison.net"
    "https://steipete.me"
    "https://writing.aref.vc"
    "https://dbreunig.com"
    "https://arpitbhayani.me"
    "https://cognition.ai"
    "https://cognition.ai/blog/1"
    "https://factory.ai/news"
    "https://docs.langchain.com"
    "https://evanhahn.com"
    "https://every.to"
    "https://leerob.com"
    "https://www.newsletter.swirlai.com"
    "https://philschmid.de"
    "https://seangoedecke.com"
    "https://vtrivedy.com"
    "https://speiss.dev"
    "https://humanlayer.dev"
    "https://josh.ing"
    "https://cartesia.ai/blog"
    "https://fsck.com"
    "https://manus.im/blog"
    "https://www.braintrust.dev/blog"
)

# Output file
output="rss_blogs.txt"
> "$output"

# Function to check for RSS feed
check_rss() {
    local url=$1
    local base_url=$(echo "$url" | sed -E 's#(https?://[^/]+).*#\1#')
    
    # Try common RSS paths
    local patterns=("feed.rss" "feed.xml" "rss" "rss.xml" "feed" "feeds" "index.xml")
    
    for pattern in "${patterns[@]}"; do
        local test_url="$base_url/$pattern"
        if curl -s -I -m 5 "$test_url" 2>/dev/null | grep -q "200\|application/rss"; then
            echo "$url → $test_url"
            return 0
        fi
    done
    
    # If no pattern matched, return the URL with no RSS found
    echo "$url → NOT FOUND"
    return 1
}

# Check each blog
for blog in "${blogs[@]}"; do
    result=$(check_rss "$blog")
    echo "$result" >> "$output"
    echo "$result"
done

echo ""
echo "Results saved to $output"
