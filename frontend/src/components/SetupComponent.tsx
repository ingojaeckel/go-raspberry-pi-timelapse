import React, { useState, useEffect, ChangeEvent } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { SettingsResponse } from '../models/response'
import { ButtonGroup, Button, Select, MenuItem, Typography, TextField, Grid } from '@material-ui/core';

export default function SetupComponent() {
  const [state, setState] = useState<SettingsResponse>({
    // Initial values shown while current config is loaded async
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
          updateForm(resp.data);
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

  // TODO replace this with proper binding
  function updateForm(data: SettingsResponse) {
    const fieldSecondsBetweenCaptures = document.getElementById("tfSecondsBetweenCaptures")! as HTMLInputElement
    const fieldOffset = document.getElementById("tfOffset")! as HTMLInputElement
    const fieldRotation = document.getElementById("tfRotation")! as HTMLInputElement
    const fieldQuality = document.getElementById("tfQuality")! as HTMLInputElement

    fieldSecondsBetweenCaptures.value = (data.SecondsBetweenCaptures / 60).toString();
    fieldOffset.value = data.OffsetWithinHour.toString();
    fieldRotation.value = data.RotateBy.toString();
    fieldQuality.value = data.Quality.toString();
  }

  const handleOffsetChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { OffsetWithinHour: parseInt(e.target.value) as number }));
  const handleQualityChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { Quality: parseInt(e.target.value) as number }));
  const handleRotationChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { RotateBy: parseInt(e.target.value) as number }));
  const handleTimeBetweenCapturesChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { SecondsBetweenCaptures: (parseInt(e.target.value) as number) * 60 }));

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
          <TextField id="tfSecondsBetweenCaptures" type="number" onChange={handleTimeBetweenCapturesChanged} defaultValue={state.SecondsBetweenCaptures/60} inputProps={{ inputMode: 'numeric', pattern: '[0-9]+' }} />
        </Grid>
        <Grid item xs={6}>
          <Typography gutterBottom>Delay within the hour before the first capture or -1 to disable (minutes):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField id="tfOffset" type="number" defaultValue={state.OffsetWithinHour} onChange={handleOffsetChanged} inputProps={{ inputMode: 'numeric', pattern: '-?[0-9]+' }} />
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
          <Typography gutterBottom>Rotation (degrees):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField id="tfRotation" type="number" defaultValue={state.RotateBy} onChange={handleRotationChanged} inputProps={{ inputMode: 'numeric', pattern: '[0-9]+' }} />
        </Grid>
        <Grid item xs={6}>
          <Typography gutterBottom>Photo Quality (0..100%):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField id="tfQuality" type="number" defaultValue={state.Quality} onChange={handleQualityChanged} inputProps={{ inputMode: 'numeric', pattern: '[0-9]+' }} />
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
