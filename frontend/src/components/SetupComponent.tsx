import React, { useState, useEffect, ChangeEvent } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { SettingsResponse } from '../models/response'
import { ButtonGroup, Button, Select, MenuItem, Typography, TextField, Grid } from '@material-ui/core';

export default function SetupComponent() {
  const [state, setState] = useState<SettingsResponse>({
    SecondsBetweenCaptures:  60,
    OffsetWithinHour:        0,
    PhotoResolutionWidth:    3280,
    PhotoResolutionHeight:   2464,
    PreviewResolutionWidth:  640,
    PreviewResolutionHeight: 480,
    RotateBy:                0,
    ResolutionSetting:       0,
    Quality:                 100,
    DebugEnabled:            false,
  });

  useEffect(() => {
    console.log("Getting previous configuration...");
    axios
      .get<SettingsResponse>(BaseUrl + "/configuration")
      .then(resp => {
        if (resp.data) {
          setState(resp.data);
        }
      });
  }, []);



  const handleRestartClicked = () => {
    console.log("restart clicked");
    axios.get(BaseUrl + "/admin/restart").then(() => console.log("restart initiated"));
  };

  const handleShutdownClicked = () => {
    console.log("shutdown clicked");
    axios.get(BaseUrl + "/admin/shutdown").then(() => console.log("shutdown initiated"));
  };

  function handleSaveSettingsClicked() {
    axios
      .post<SettingsResponse>(BaseUrl + "/configuration", state)
      .then(resp => {
        console.log("Settings updated to ", resp.data);
        setState(resp.data);
      });
  };

  const handleOffsetChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { OffsetWithinHour: parseInt(e.target.value) as number }));
  const handleQualityChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { Quality: parseInt(e.target.value) as number }));
  const handleRotationChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { RotateBy: parseInt(e.target.value) as number }));
  const handleTimeBetweenCapturesChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { SecondsBetweenCaptures: (parseInt(e.target.value) as number) * 60 }));

  // TODO Fix bug: Values only update on save
  return (
    <React.Fragment>
      <div>
        <Typography variant="h4" component="h4">Settings</Typography>
      </div>
      <Grid container spacing={2}>
        <Grid item xs={6}>
          <Typography gutterBottom>Time between captures (minutes):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField type="number" onChange={handleTimeBetweenCapturesChanged} defaultValue={state.SecondsBetweenCaptures/60} inputProps={{ inputMode: 'numeric', pattern: '[0-9]+' }}  />
        </Grid>
        <Grid item xs={6}>
          <Typography gutterBottom>Delay within value hour before first capture (minutes): {state.OffsetWithinHour}</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField type="number" defaultValue={state.OffsetWithinHour} onChange={handleOffsetChanged} />
        </Grid>
        <Grid item xs={6}>
          <Typography gutterBottom>Photo Resolution (pixels)</Typography>
        </Grid>
        <Grid item xs={6}>
          <Select defaultValue={0}>
            <MenuItem value={0}>3280x2464</MenuItem>
          </Select>
        </Grid>
        <Grid item xs={6}>
          <Typography gutterBottom>Rotation (degrees): {state.RotateBy}</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField type="number" defaultValue={state.RotateBy} onChange={handleRotationChanged} />
        </Grid>
        <Grid item xs={6}>
          <Typography gutterBottom>Photo Quality (0..100%): {state.Quality}</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField type="number" defaultValue={state.Quality} onChange={handleQualityChanged} />
        </Grid>
        <Grid item xs={6}>
        </Grid>
      </Grid>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleSaveSettingsClicked}>Save</Button>
        <Button onClick={handleRestartClicked}>Restart</Button>
        <Button onClick={handleShutdownClicked}>Shutdown</Button>
      </ButtonGroup>
    </React.Fragment>
  );
}
