package main

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

func Test_searchID(t *testing.T) {
	time1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2023, 1, 1, 12, 10, 0, 0, time.UTC)
	time3 := time.Date(2023, 1, 1, 12, 20, 0, 0, time.UTC)

	type args struct {
		items []item
		id    string
	}
	type want struct {
		index int
		item  item
		found bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "item with id 1, exist",
			args: args{
				items: []item{
					{ID: "1", Name: "Item 1", Timestamp: time1},
					{ID: "2", Name: "Item 2", Timestamp: time2},
					{ID: "3", Name: "Item 3", Timestamp: time3},
				},
				id: "1",
			},
			want: want{
				index: 0,
				item:  item{ID: "1", Name: "Item 1", Timestamp: time1},
				found: true,
			},
		},
		{
			name: "item with id 2, exist",
			args: args{
				items: []item{
					{ID: "1", Name: "Item 1", Timestamp: time1},
					{ID: "2", Name: "Item 2", Timestamp: time2},
					{ID: "3", Name: "Item 3", Timestamp: time3},
				},
				id: "2",
			},
			want: want{
				index: 1,
				item:  item{ID: "2", Name: "Item 2", Timestamp: time2},
				found: true,
			},
		},
		{
			name: "item with id 5, not exist",
			args: args{
				items: []item{
					{ID: "1", Name: "Item 1", Timestamp: time1},
					{ID: "2", Name: "Item 2", Timestamp: time2},
					{ID: "3", Name: "Item 3", Timestamp: time3},
				},
				id: "5",
			},
			want: want{
				index: -1,
				item:  item{},
				found: false,
			},
		},
		{
			name: "empty slice",
			args: args{
				items: []item{},
				id:    "1",
			},
			want: want{
				index: -1,
				item:  item{},
				found: false,
			},
		},
		{
			name: "empty id",
			args: args{
				items: []item{
					{ID: "1", Name: "Item 1", Timestamp: time1},
					{ID: "2", Name: "Item 2", Timestamp: time2},
				},
				id: "",
			},
			want: want{
				index: -1,
				item:  item{},
				found: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex, gotItem, gotFound := searchID(tt.args.items, tt.args.id)
			if gotIndex != tt.want.index {
				t.Errorf("searchID() gotIndex = %v, want %v", gotIndex, tt.want.index)
			}
			if !reflect.DeepEqual(gotItem, tt.want.item) {
				t.Errorf("searchID() gotItem = %v, want %v", gotItem, tt.want.item)
			}
			if gotFound != tt.want.found {
				t.Errorf("searchID() gotFound = %v, want %v", gotFound, tt.want.found)
			}
		})
	}
}

func Test_parseQuery(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    QueryParams
		wantErr bool
	}{
		{
			name: "Valid JSON format and ID sort",
			url:  "/path?format=json&sort=id",
			want: QueryParams{
				Format: "json",
				Sort:   "id",
			},
			wantErr: false,
		},
		{
			name: "Valid YAML format and timestamp sort",
			url:  "/path?format=yaml&sort=timestamp",
			want: QueryParams{
				Format: "yaml",
				Sort:   "timestamp",
			},
			wantErr: false,
		},
		{
			name: "Empty query parameters",
			url:  "/path",
			want: QueryParams{
				Format: "",
				Sort:   "",
			},
			wantErr: false,
		},
		{
			name:    "Invalid format parameter",
			url:     "/path?format=invalid&sort=id",
			want:    QueryParams{},
			wantErr: true,
		},
		{
			name:    "Invalid sort parameter",
			url:     "/path?format=json&sort=invalid",
			want:    QueryParams{},
			wantErr: true,
		},
		{
			name:    "Both parameters invalid",
			url:     "/path?format=invalid&sort=invalid",
			want:    QueryParams{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			got, err := parseQuery(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseQuery() got = %v, want %v", got, tt.want)
			}
		})
	}
}
