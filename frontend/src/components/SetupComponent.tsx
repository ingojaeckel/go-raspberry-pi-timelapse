import React, { useState, useEffect, ChangeEvent } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { SettingsResponse } from '../models/response'
import { ButtonGroup, Button, Select, MenuItem, Slider, Typography } from '@material-ui/core';

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

  const handleOffsetChanged = (_event: ChangeEvent<{}>, value: number | number[]) => setState(Object.assign(state, { OffsetWithinHour: value as number }));
  const handleQualityChanged = (_event: ChangeEvent<{}>, value: number | number[]) => setState(Object.assign(state, { Quality: value as number }));
  const handleRotationChanged = (_event: ChangeEvent<{}>, value: number | number[]) => setState(Object.assign(state, { RotateBy: value as number }));
  const handleTimeBetweenCapturesChanged = (_event: ChangeEvent<{}>, value: number | number[]) => setState(Object.assign(state, { SecondsBetweenCaptures: (value as number) * 60 }));

  // TODO Fix bug: Values only update on save
  return (
    <React.Fragment>
      <div>
        <Typography variant="h4" component="h4">Settings</Typography>
      </div>
      <div>
        <Typography gutterBottom>Time between captures (minutes): {state.SecondsBetweenCaptures/60}</Typography>
        <Slider valueLabelDisplay="auto" step={1} marks min={1} max={30} value={state.SecondsBetweenCaptures/60} onChange={handleTimeBetweenCapturesChanged} />
        <Typography gutterBottom>Delay within value hour before first capture (minutes): {state.OffsetWithinHour}</Typography>
        <Slider valueLabelDisplay="auto" step={5} marks min={0} max={30} value={state.OffsetWithinHour} onChange={handleOffsetChanged} />
        <Typography gutterBottom>Photo Resolution (pixels)</Typography>
        <Select defaultValue={0}>
          <MenuItem value={0}>3280x2464</MenuItem>
        </Select>
        <Typography gutterBottom>Rotation (degrees): {state.RotateBy}</Typography>
        <Slider valueLabelDisplay="auto" step={180} marks min={0} max={180} value={state.RotateBy} onChange={handleRotationChanged} />
        <Typography gutterBottom>Photo Quality (0..100%): {state.Quality}</Typography>
        <Slider valueLabelDisplay="auto" step={5} marks min={0} max={100} value={state.Quality} onChange={handleQualityChanged} />
      </div>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleSaveSettingsClicked}>Save</Button>
        <Button onClick={handleRestartClicked}>Restart</Button>
        <Button onClick={handleShutdownClicked}>Shutdown</Button>
      </ButtonGroup>
    </React.Fragment>
  );
}
