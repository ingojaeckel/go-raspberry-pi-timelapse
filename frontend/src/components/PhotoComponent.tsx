import React, { useState, useEffect } from 'react';
import { Typography, Button, ButtonGroup } from '@material-ui/core';
import axios from 'axios';
import { DataGrid, ColDef, RowData, RowId, SelectionModelChangeParams, CellParams } from '@material-ui/data-grid';
import { PhotosResponse } from '../models/response';
import { BaseUrl } from '../conf/config'

export interface PhotosRowData {
  Photos: RowData[],
  Selected: RowId[],
  SelectedFilesParameter: string,
}

const columns: ColDef[] = [
  { field: 'fileName', headerName: 'Name', width: 300, renderCell: (p: CellParams) => (<a href={BaseUrl + "/file/" + p.value}>{p.value}</a> ) },
  { field: 'fileCreateTime', headerName: 'Created At', width: 300 },
  { field: 'fileSizeBytes', headerName: 'Size', width: 100 },
];

export default function PhotosComponent() {
  const [state, setState] = useState<PhotosRowData>({
    Photos: [],
    Selected: [],
    SelectedFilesParameter: "",
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
          SelectedFilesParameter: "",
        });
      }
    });
  }

  useEffect(() => {
    getPhotos()
  }, []);

  const handleSelectionModelChanged = (params: SelectionModelChangeParams) => {
    console.log("selection changed: ", params);

    var link = "";
    params.selectionModel.forEach(selected => {
      let selectedPhoto = state.Photos.find(e => e.id === selected);
      if (selectedPhoto) {
        let selectedPhotoFilename = selectedPhoto.fileName;
        link += "f=" + selectedPhotoFilename + "&";
      }
    });

    setState({
      Photos: state.Photos,
      Selected: params.selectionModel,
      SelectedFilesParameter: link,
    });
  };

  const handleRefreshClicked = () => {
    console.log("refresh clicked");
    getPhotos();
  };

  return (
    <React.Fragment>
      <Typography variant="h4" component="h4">Photos</Typography>
      <div>
        <Typography variant="h5" component="h5">Preview Capture</Typography>
        <img src={BaseUrl+"/capture"} alt="preview" />
      </div>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleRefreshClicked}>Refresh list of pictures</Button>
      </ButtonGroup>
      <ul>
        <li><a href={BaseUrl + "/archive/zip"}>Download all</a></li>
        <li><a href={BaseUrl + "/archive/zip?" + state.SelectedFilesParameter}>Download selected ({state.Selected.length})</a></li>
        <li>
          <small>
            <a href={BaseUrl + "/files/delete?" + state.SelectedFilesParameter}>Delete selected ({state.Selected.length})</a>
          </small>
        </li>
      </ul>
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
