import type { Action } from "@reduxjs/toolkit";
import { drawingApi } from "./drawingApi";

export const drawingErrorLabels = (action: Action): string | undefined => {
	if (drawingApi.endpoints.getDrawingRepositories.matchRejected(action)) {
		return "Failed to load repositories";
	}
	if (drawingApi.endpoints.getDrawingLists.matchRejected(action)) {
		return "Failed to load drawing list";
	}
	if (drawingApi.endpoints.getDrawing.matchRejected(action)) {
		return "Failed to open drawing";
	}
	if (drawingApi.endpoints.createDrawing.matchRejected(action)) {
		return "Failed to create drawing";
	}
	if (drawingApi.endpoints.saveDrawing.matchRejected(action)) {
		return "Failed to save drawing";
	}
	if (drawingApi.endpoints.deleteDrawings.matchRejected(action)) {
		return "Failed to delete drawing(s)";
	}
	return undefined;
};
