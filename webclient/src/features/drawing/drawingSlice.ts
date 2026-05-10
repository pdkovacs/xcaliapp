import { castDraft } from "immer";
import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import { emptyArray } from "../../utils/empty-array";
import type { Drawing } from "./drawingApi";

const emptyDrawing: Drawing = {
	id: "",
	title: "",
	elements: emptyArray,
	repo: {
		name: "",
		label: ""
	}
};

export interface DrawingSliceState {
	savedDrawing: Drawing;
}

const initialState: DrawingSliceState = {
	savedDrawing: emptyDrawing
};

export const drawingSlice = createSlice({
	name: "drawing",
	initialState,
	reducers: {
		clearCanvas: state => {
			state.savedDrawing = castDraft(emptyDrawing);
		},
		setSavedDrawing: (state, action: PayloadAction<Drawing>) => {
			state.savedDrawing = castDraft(action.payload);
		}
	},
	selectors: {
		selectSavedDrawing: state => state.savedDrawing
	}
});

export const { clearCanvas, setSavedDrawing } = drawingSlice.actions;
export const { selectSavedDrawing } = drawingSlice.selectors;
