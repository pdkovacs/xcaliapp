import { Backdrop, CircularProgress } from "@mui/material";
import { useEffect, useState } from "react";

export const LoadingBackdrop = ({ open, onCancel }: { open: boolean; onCancel: () => void }) => {
	const [visible, setVisible] = useState(false);

	useEffect(() => {
		if (!open) {
			setVisible(false);
			return;
		}
		const t = setTimeout(() => setVisible(true), 200);
		return () => clearTimeout(t);
	}, [open]);

	return (
		<Backdrop open={open && visible} onClick={onCancel} sx={{ zIndex: theme => theme.zIndex.modal + 1 }}>
			<CircularProgress />
		</Backdrop>
	);
};

// TODO: this smells like a pre-mature abstraction
export interface GatedGlobalResourceLoadingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
}