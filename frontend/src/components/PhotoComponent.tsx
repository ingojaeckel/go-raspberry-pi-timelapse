import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Typography, Button, ButtonGroup } from '@material-ui/core';
import { DataGrid, ColDef, RowData, RowId, SelectionModelChangeParams, CellParams } from '@material-ui/data-grid';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import { PhotosResponse } from '../models/response';
import { BaseUrl } from '../conf/config'

export interface PhotosRowData {
  ShowDeletionDialog: boolean,
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
    ShowDeletionDialog: false,
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
          ShowDeletionDialog: false,
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
      ShowDeletionDialog: false,
      Photos: state.Photos,
      Selected: params.selectionModel,
      SelectedFilesParameter: link,
    });
  };

  const handleRefreshClicked = () => getPhotos();
  const deletePhotosClicked = () => {
    if (state.Selected.length > 0) {
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
      <Typography variant="h4" component="h4">Photos</Typography>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleRefreshClicked}>Refresh</Button>
        <Button onClick={deletePhotosClicked}>Delete selected ({state.Selected.length})</Button>
      </ButtonGroup>
      <ul>
        <li><a href={BaseUrl + "/archive/zip"}>Download all</a></li>
        <li><a href={BaseUrl + "/archive/zip?" + state.SelectedFilesParameter}>Download selected ({state.Selected.length})</a></li>
      </ul>
      <div style={{ height: 500, width: '100%' }}>
        <DataGrid
          rows={state.Photos}
          columns={columns}
          checkboxSelection={true}
          disableSelectionOnClick={true}
          onSelectionModelChange={handleSelectionModelChanged} />
      </div>
      <Dialog open={state.ShowDeletionDialog} onClose={handleDeletionCancelled} aria-describedby="alert-dialog-description">
        <DialogContent>
          <DialogContentText id="alert-dialog-description">Are you sure you want to delete the selected {state.Selected.length} files?</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeletionCancelled} color="primary">Cancel</Button>
          <Button onClick={handleDeletionConfirmed} color="primary" autoFocus>Delete</Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  );
}
