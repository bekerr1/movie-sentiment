# Problem Statement

Our user sentiment team wants to analyze reviews from unhappy users to figure out how to improve their experience. This is currently tricky because the reviews are spread out over many servers.

The user sentiment team has put in a feature request for you to combine all the negative user reviews so that the reviews are easier for them to analyze.

Example:

```
On server-1

Filename - movie-foo.txt:
user1: This should have been in free movie downloads
user2: Great movie
user3: Most boring ever
http://server-2/movie-foo.txt   -> Link to next file 
```


```
On server-2

Filename - movie-foo.txt:
user4: Amazing movie
user5: Did not enjoy it. Pacing was off.
```

The reviews are stored in one or more text files. Each file contains reviews for a single movie. The last line of the file may contain a URL to where to find more reviews for that same movie.

The user sentiment team will provide the URLs of where to find the first review file for each movie. They will also provide a list of negative sentiment phrases. A review is considered negative if it contains one of these phrases (case insensitive). 

```
reviews = ["http://server-1/movie-foo.txt", "http://server-1/movie-bar.txt"]

negativePhrases = [
  "free movie downloads",
  "Movie is disappointing",
  "Pacing was off",
  "Boring",
]
```
Your task is to build a system that provides aggregated negative reviews - one output file per movie.

The company has A LOT of users and A LOT of movies-- there are many movie files each of which contains many reviews! As you design your solution keep in mind that this is a production system.

# Approach

- Files/Servers can be crawled in an iterative manner.
- File sizes can be large, we need to account for memory consumption when
  downloading/streaming content over the network.
- We can consider parallel processing of movies since they have no dependancy
  on eachother
- Simple substring matching on negative sentiment is OK for string processing

# Edge Cases

The following are some cases that make this problem a bit harder to "just do". Some of these cases might have better ways to handle them than I did, and maybe taht would negate their complexity.

## Handling Chunks when newline characters are split

When reading file contents in chunks, its possible to read in the middle of some current line and new line, something like... 

```
....
READ1{user1: this move was} READ2{ boring\n"
user2: I loved it\n}"
....
```

From the above, READ1 ends in the middle of the line. As such, we have to save this partial line and prepend it to the next chunk read (READ2).
Then continue processing as normal.

## The last line may not be a URL 

The last line of the file may not be a URL. In this case, we need to make sure we don't try to fetch a URL that doesn't exist.

## Handle negative sentiment memory efficiently

Even though we read files efficiently in chunks, it also may be the case we need to incrementally write the negative sentiments to disk to ensure we dont run out of memory.

# Build/Run 

There are pre-generated test artifacts to use. All you have to do is build the binary and run docker-compose to start the simulation.
There is also some validation within the docker container running the gathering.

```
make build 
make up
```

# NOTE

Some stuff in here is hardcoded for convinience and I ack there are probably more "elegant" ways to do things as far as running/testing. Through I'm generally ok with the code, if anything looks especially off, please call it out.
