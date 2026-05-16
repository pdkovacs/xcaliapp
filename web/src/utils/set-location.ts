import { isNil } from "lodash";
import { QualifiedDrawingId, qualifiedDrawingIdToString } from "../features/drawing/drawingApi";

export const setLocation = (qualifiedDrawingId: QualifiedDrawingId | null) => {
	if (!isNil(qualifiedDrawingId)) {
		const drawingId = qualifiedDrawingIdToString(qualifiedDrawingId);
		const desiredPathname = `/drawings/${drawingId}`;
		if (window.location.pathname !== desiredPathname) {
			window.history.pushState({}, drawingId, desiredPathname);
		}
	} else {
		if (window.location.pathname !== "") {
			window.history.pushState({}, "", "/");
		}
	}
};
