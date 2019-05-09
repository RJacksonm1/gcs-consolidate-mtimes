package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func fiddleMtimes(ctx context.Context, objAttrs *storage.ObjectAttrs, client *storage.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	// Skip if file was not rsynced via gsutil
	rsyncMtime, rsyncMtimeExists := objAttrs.Metadata["goog-reserved-file-mtime"]
	if !rsyncMtimeExists {
		return
	}

	// Skip if gcsfuse is already tracking this object's mtime
	if _, gcsfuseMtimeExists := objAttrs.Metadata["gcsfuse_mtime"]; gcsfuseMtimeExists {
		return
	}

	intMtime, err := strconv.ParseInt(rsyncMtime, 10, 64)
	if err != nil {
		fmt.Println("Really wonky mtime found", objAttrs.Name, rsyncMtime)
	}
	mtime := time.Unix(intMtime, 0)
	formattedTime := mtime.UTC().Format(time.RFC3339Nano)

	fmt.Printf("Setting the gcsfuse mtime of %s to %s\n", objAttrs.Name, formattedTime)
	objAttrs, err = client.Bucket(objAttrs.Bucket).Object(objAttrs.Name).Update(ctx, storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"gcsfuse_mtime": formattedTime,
		},
	})
	if err != nil {
		fmt.Println("Couldn't set mtime", objAttrs.Name, formattedTime)
	}
}

func main() {
	// Validate input
	if len(os.Args) != 2 {
		log.Fatalf("Invalid number of arguments.")
	}
	bucketName := os.Args[1]

	// Set up Google Cloud Storage wibbly wobs
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Find and fiddle objects
	var wg sync.WaitGroup
	it := client.Bucket(bucketName).Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}

		wg.Add(1)
		go fiddleMtimes(ctx, objAttrs, client, &wg)
	}

	wg.Wait()
}
