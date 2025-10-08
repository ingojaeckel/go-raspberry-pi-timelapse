import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { DataGrid, GridColDef, GridRowId, GridRenderCellParams, GridRowSelectionModel } from '@mui/x-data-grid';
import { Button, ButtonGroup, Dialog, DialogActions, DialogContent, DialogContentText } from '@mui/material';
import { BaseUrl } from '../conf/config'
import { PhotosResponse } from '../models/response';

export interface PhotosRowData {
  ShowDeletionDialog: boolean,
  Photos: any[],
  Selected: GridRowSelectionModel,
  SelectedFilesParameter: string,
}

const columns: GridColDef[] = [
  { field: 'fileName', headerName: 'Name', width: 300, renderCell: (p: GridRenderCellParams) => (<a href={BaseUrl + "/file/" + p.value}>{p.value}</a> ) },
  { field: 'fileCreateTime', headerName: 'Created At', width: 300 },
  { field: 'fileSizeBytes', headerName: 'Size', width: 100 },
];

export default function PhotosComponent() {
  const [state, setState] = useState<PhotosRowData>({
    ShowDeletionDialog: false,
    Photos: [],
    Selected: { type: 'include', ids: new Set<GridRowId>() },
    SelectedFilesParameter: "",
  });

const getPhotos = () => {
  console.log("BaseUrl: ", BaseUrl)
  axios
    .get<PhotosResponse>(BaseUrl + "/photos")
    .then(resp => {
      // After receiving a response, map the PhotosResponse to RowData[] which can be displayed in the data grid.
      if (resp.data && resp.data.Photos) {
        var rows: any[] = [];

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
          ShowDeletionDialog: false,
          Photos: rows,
          Selected: { type: 'include', ids: new Set<GridRowId>() },
          SelectedFilesParameter: "",
        });
      }
    });
  }

  useEffect(() => {
    getPhotos()
  }, []);

  const handleSelectionModelChanged = (selectionModel: GridRowSelectionModel) => {
    console.log("selection changed: ", selectionModel);

    var link = "";
    selectionModel.ids.forEach(selected => {
      let selectedPhoto = state.Photos.find(e => e.id === selected);
      if (selectedPhoto) {
        let selectedPhotoFilename = selectedPhoto.fileName;
        link += "f=" + selectedPhotoFilename + "&";
      }
    });

    setState({
      ShowDeletionDialog: false,
      Photos: state.Photos,
      Selected: selectionModel,
      SelectedFilesParameter: link,
    });
  };

  const handleRefreshClicked = () => getPhotos();
  const deletePhotosClicked = () => {
    if (state.Selected.ids.size > 0) {
      setState({
        ShowDeletionDialog: true,
        Photos: state.Photos,
        Selected: state.Selected,
        SelectedFilesParameter: state.SelectedFilesParameter,
      })
    }
  };
  const handleDeletionCancelled = () => {
    setState({
      ShowDeletionDialog: false,
      Photos: state.Photos,
      Selected: state.Selected,
      SelectedFilesParameter: state.SelectedFilesParameter,
    })
  };
  const handleDeletionConfirmed = () => {
    axios
    .get(BaseUrl + "/file/delete?" + state.SelectedFilesParameter)
    .then(_resp => {
      setState({
        ShowDeletionDialog: false,
        Photos: state.Photos,
        Selected: state.Selected,
        SelectedFilesParameter: state.SelectedFilesParameter,
      });
      getPhotos();
    });
  };

  return (
    <React.Fragment>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleRefreshClicked}>Refresh</Button>
        <Button onClick={deletePhotosClicked}>Delete selected ({state.Selected.ids.size})</Button>
      </ButtonGroup>
      <ul>
        <li><a href={BaseUrl + "/archive/zip"}>Download all (zip)</a></li>
        <li><a href={BaseUrl + "/archive/tar"}>Download all (tar)</a></li>
        <li><a href={BaseUrl + "/archive/zip?" + state.SelectedFilesParameter}>Download selected ({state.Selected.ids.size})</a></li>
      </ul>
      <div style={{ height: 500, width: '100%' }}>
        <DataGrid
          rows={state.Photos}
          columns={columns}
          checkboxSelection={true}
          disableRowSelectionOnClick={true}
          onRowSelectionModelChange={handleSelectionModelChanged} />
      </div>
      <Dialog open={state.ShowDeletionDialog} onClose={handleDeletionCancelled} aria-describedby="alert-dialog-description">
        <DialogContent>
          <DialogContentText id="alert-dialog-description">Are you sure you want to delete the selected {state.Selected.ids.size} files?</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeletionCancelled} color="primary">Cancel</Button>
          <Button onClick={handleDeletionConfirmed} color="primary" autoFocus>Delete</Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  );
}
