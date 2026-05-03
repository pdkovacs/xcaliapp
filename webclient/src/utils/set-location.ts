import { isNil } from "lodash";
import { FQDrawingId, fqDrawingIdToString } from "../features/drawing/drawingAPI";

export const setLocation = (fQDrawingId: FQDrawingId | null) => {
	if (!isNil(fQDrawingId)) {
		const drawingId = fqDrawingIdToString(fQDrawingId);
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
