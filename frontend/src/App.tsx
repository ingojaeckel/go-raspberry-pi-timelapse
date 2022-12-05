import React from 'react';
import { createBrowserHistory } from "history";
import { NavLink, Router } from 'react-router-dom';
import Switcher from './Switch';
import axios from 'axios';
import { Avatar, Container, CssBaseline, Grid, List, ListItem, ListItemAvatar } from '@mui/material';
import { Home, PhotoCamera, Timeline, Settings, Description } from '@mui/icons-material'

// Register default request interceptors
axios.interceptors.request.use(request => {
  console.log('Starting Request', JSON.stringify(request, null, 2))
  return request
})

const customHistory = createBrowserHistory();

function App() {
  return (
    <Router history={customHistory}>
      <React.Fragment>
        <CssBaseline />
        <Container fixed>
          <Grid container spacing={0}>
            <Grid item xs={2} alignItems="center">
              <List disablePadding>
                <ListItem to={`${process.env.PUBLIC_URL}/`} component={NavLink} disableGutters>
                  <ListItemAvatar>
                    <Avatar>
                      <Home />
                    </Avatar>
                  </ListItemAvatar>
                </ListItem>
                <ListItem to={`${process.env.PUBLIC_URL}/preview`} component={NavLink} disableGutters>
                  <ListItemAvatar>
                    <Avatar>
                      <PhotoCamera />
                    </Avatar>
                  </ListItemAvatar>
                </ListItem>
                <ListItem to={`${process.env.PUBLIC_URL}/monitoring`} component={NavLink} disableGutters>
                  <ListItemAvatar>
                    <Avatar>
                      <Timeline />
                    </Avatar>
                  </ListItemAvatar>
                </ListItem>
                <ListItem to={`${process.env.PUBLIC_URL}/settings`} component={NavLink} disableGutters>
                  <ListItemAvatar>
                    <Avatar>
                      <Settings />
                    </Avatar>
                  </ListItemAvatar>
                </ListItem>
                <ListItem to={`${process.env.PUBLIC_URL}/logs`} component={NavLink} disableGutters>
                  <ListItemAvatar>
                    <Avatar>
                      <Description />
                    </Avatar>
                  </ListItemAvatar>
                </ListItem>
              </List>
            </Grid>
            <Grid item xs={10}>
              <Switcher />
            </Grid>
            <Grid item xs={12}>
              <div className="footer">version: <a href={"https://github.com/ingojaeckel/go-raspberry-pi-timelapse/commit/" + process.env.REACT_APP_GIT_SHA}>{process.env.REACT_APP_GIT_SHA}</a></div>
            </Grid>
          </Grid>
        </Container>
      </React.Fragment>
    </Router>
  );
}

export default App;
