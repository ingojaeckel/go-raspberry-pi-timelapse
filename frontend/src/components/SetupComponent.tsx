import React from 'react';
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';

import { ButtonGroup, Button, FormControl, InputLabel, Select, MenuItem, FormHelperText } from '@material-ui/core';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    formControl: {
      margin: theme.spacing(1),
      minWidth: 120,
    },
    selectEmpty: {
      marginTop: theme.spacing(2),
    },
  }),
);

export default function SetupComponent() {
  const classes = useStyles();

  return (
    <React.Fragment>
      <div>
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Time between captures (minutes)</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper">
            <MenuItem value={1}>1</MenuItem>
            <MenuItem value={10}>10</MenuItem>
            <MenuItem value={20}>20</MenuItem>
            <MenuItem value={30}>30</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl><br />
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Offset before first capture (minutes)</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper">
            <MenuItem value={0}>0</MenuItem>
            <MenuItem value={5}>5</MenuItem>
            <MenuItem value={10}>10</MenuItem>
            <MenuItem value={15}>15</MenuItem>
            <MenuItem value={20}>20</MenuItem>
            <MenuItem value={25}>25</MenuItem>
            <MenuItem value={30}>30</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl><br />
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Resolution</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper">
            <MenuItem>3280x2464</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl><br />
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Rotation (degrees)</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper">
            <MenuItem>0</MenuItem>
            <MenuItem>180</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl>
      </div>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button>Restart</Button>
        <Button>Shutdown</Button>
        <Button>Delete Photos</Button>
      </ButtonGroup>
    </React.Fragment>
  );
}
