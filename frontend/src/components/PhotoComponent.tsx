import React from 'react';
import PhotoListComponent from './PhotoListComponent';
import { Typography, Button } from '@material-ui/core';

export default function PhotosComponent() {
  return (
    <React.Fragment>
      <Typography variant="h5" component="h5">Preview</Typography>
      <Typography variant="body1">Current photo goes here.</Typography>
      <Button variant="contained">Download</Button>
      <PhotoListComponent />
    </React.Fragment>
  );
}
