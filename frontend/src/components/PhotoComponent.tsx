import React, { useState, useEffect } from 'react';
import { Typography, Button, ButtonGroup } from '@material-ui/core';
import axios from 'axios';
import { DataGrid, ColDef, RowData, RowId, SelectionModelChangeParams } from '@material-ui/data-grid';
import { PhotosResponse } from '../models/response';
import { BaseUrl } from '../conf/config'

export interface PhotosRowData {
  Photos: RowData[],
  Selected: RowId[],
  ArchiveFilterParameter: string,
}

const columns: ColDef[] = [
  { field: 'fileName', headerName: 'Name', width: 300 },
  { field: 'fileCreateTime', headerName: 'Created At', width: 300 },
  { field: 'fileSizeBytes', headerName: 'Size', width: 100 },
];

export default function PhotosComponent() {
  const [state, setState] = useState<PhotosRowData>({
    Photos: [],
    Selected: [],
    ArchiveFilterParameter: "",
  });

const getPhotos = () => {
  console.log("BaseUrl: ", BaseUrl)
  axios
    .get<PhotosResponse>(BaseUrl + "/photos")
    .then(resp => {
      // After receiving a response, map the PhotosResponse to RowData[] which can be displayed in the data grid.
      if (resp.data) {
        var rows: RowData[] = [];

        for (var i=0; i<resp.data.Photos.length; i++) {
          let photo = resp.data.Photos[i];
          rows.push({
            id: photo.Name,
            isSelected: false,
            fileName: photo.Name,
            fileCreateTime: photo.ModTime,
            fileSizeBytes: photo.Size})
        }

        setState({
          Photos: rows,
          Selected: [],
          ArchiveFilterParameter: "",
        });
      }
    });
  }

  useEffect(() => {
    getPhotos()
  }, []);

  const handleSelectionModelChanged = (selectionModel: SelectionModelChangeParams) => {
    console.log("selection changed: ", selectionModel);

    var link = "";
    for (var i=0; i<selectionModel.selectionModel.length; i++) {
      let selectedPhoto = state.Photos.find(e => e.id == selectionModel.selectionModel[i]);
      if (selectedPhoto) {
        let selectedPhotoFilename = selectedPhoto.fileName;
        link += "f=" + selectedPhotoFilename + "&";
      }
    }

    setState({
      Photos: state.Photos,
      Selected: selectionModel.selectionModel,
      ArchiveFilterParameter: link,
    });
  };

  const handleRefreshClicked = () => {
    console.log("refresh clicked");
    getPhotos();
  };
  const handleDeleteClicked = () => { console.log("delete clicked: ", state.Selected) };

  return (
    <React.Fragment>
      <Typography variant="h5" component="h5">Preview</Typography>
      <Typography variant="body1">
        <img src={BaseUrl+"/capture"} alt="preview" />
      </Typography>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleRefreshClicked}>Refresh list of pictures</Button>
        <Button onClick={handleDeleteClicked}>Delete selected pictures</Button>
      </ButtonGroup>
      <a href={BaseUrl + "/archive/zip?" + state.ArchiveFilterParameter}>Download selected pictures</a>
      <div style={{ height: 500, width: '100%' }}>
        <DataGrid
          rows={state.Photos}
          columns={columns}
          checkboxSelection={true}
          disableSelectionOnClick={true}
          onSelectionModelChange={handleSelectionModelChanged} />
        </div>
      </React.Fragment>
  );
}
