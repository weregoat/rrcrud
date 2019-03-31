// Package storage handles the functions for processing members' records in the BoltDB database
package storage

import (
	"encoding/json"
	"github.com/boltdb/bolt"
)

// Member defines the structure of the member record.
type Member struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// MembersBucket defines the bucket to use for members' records.
var MembersBucket = []byte("members")

// GetMembers returns all the members in the database.
func GetMembers(db *bolt.DB) (map[string]Member, error) {
	members := make(map[string]Member)
	err := db.View(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket(MembersBucket)
		bucket.ForEach(func(k, v []byte) error {
			member := Member{}
			err = json.Unmarshal(v, &member)
			if err == nil && len(member.ID) > 0 {
				members[member.ID] = member
			}
			return err
		})
		return err
	})
	return members, err
}

// Update replaces the record in the database with the same ID
func Update(db *bolt.DB, member Member) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(MembersBucket)
		if err == nil {
			data, err := json.Marshal(member)
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(member.ID), data)
			if err != nil {
				return err
			}

		}
		return err
	})
	return err
}

// Delete removes from the database a record with the given ID.
func Delete(db *bolt.DB, id string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(MembersBucket)
		err := bucket.Delete([]byte(id))
		return err
	})
	return err
}

// Get returns the member with the given ID from the database.
func Get(db *bolt.DB, id string) (Member, error) {
	member := Member{}
	err := db.View(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket(MembersBucket)
		data := bucket.Get([]byte(id))
		if data != nil {
			err = json.Unmarshal(data, &member)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return member, err
}

