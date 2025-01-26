import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Select } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { getDrawingContent, getDrawingList, selectDrawingList, selectDrawingListStatus } from "./drawingSlice";

import style from "./Drawing.module.css";

interface OpenDrawingDialogProps {
  readonly open: boolean;
  readonly onClose: () => void;
}

export const OpenDrawingDialog = ({ open, onClose }: OpenDrawingDialogProps) => {

  const drawingList = useAppSelector(selectDrawingList);
  const drawingListStatus = useAppSelector(selectDrawingListStatus);

  const [titleOfSelectedDrawing, setTitleOfSelectedDrawing] = useState<string>("");

  const dispatch = useAppDispatch();

  useEffect(() => {
    if (open) {
      dispatch(getDrawingList());
    }
  }, [open]);

  const handleOk = () => {
    dispatch(getDrawingContent(titleOfSelectedDrawing));
    onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
    >
      <DialogTitle>Open drawing</DialogTitle>
      <DialogContent>
        <div className={style.openSaveDrawingDialogContent}>{
          drawingListStatus === "loading"
            ? <CircularProgress />
            : drawingListStatus === "failed"
              ? <Alert severity="error">Failed to load drawing list</Alert>
              : <DrawingSelector drawingTitles={drawingList} selectedDrawing={titleOfSelectedDrawing} onChange={selection => setTitleOfSelectedDrawing(selection)}/>
        }</div>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleOk} autoFocus>OK</Button>
      </DialogActions>
    </Dialog>
  );
};

interface DrawingSelectorProps {
  readonly drawingTitles: string[];
  readonly selectedDrawing: string;
  readonly onChange: (selectedDrawing: string) => void;
}

const DrawingSelector = ({ drawingTitles: drawingTitles, selectedDrawing, onChange }:  DrawingSelectorProps) => {
  const selectedTitle = selectedDrawing || drawingTitles[0] || "";
  useEffect(() => {
    onChange(selectedTitle);
  }, [selectedTitle]);
  return (
    <Select className={style.openSaveDrawingDialogContent}
      label="Drawing"
      onChange={event => onChange(event.target.value)}
      value={selectedTitle}
    >{
        drawingTitles.map(drawingTitle => <MenuItem key={drawingTitle} value={drawingTitle}>{drawingTitle}</MenuItem>) 
    }</Select>
  );
};